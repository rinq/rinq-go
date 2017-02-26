package amqp

import (
	"context"
	"os"
	"path"

	"github.com/over-pass/overpass-go/src/overpass"
	"github.com/over-pass/overpass-go/src/overpass/amqp/internal/amqputil"
	"github.com/over-pass/overpass-go/src/overpass/amqp/internal/commandamqp"
	"github.com/over-pass/overpass-go/src/overpass/amqp/internal/notifyamqp"
	"github.com/over-pass/overpass-go/src/overpass/internal/localsession"
	"github.com/over-pass/overpass-go/src/overpass/internal/remotesession"
	"github.com/over-pass/overpass-go/src/overpass/internal/revision"
	"github.com/streadway/amqp"
)

// Dialer connects to an AMQP-based Overpass network, establishing the peer's
// unique identity on th enetwork.
type Dialer struct {
	// The minimum number of AMQP channels to keep open. If PoolSize is zero,
	// DefaultPoolSize is used.
	PoolSize uint

	// Configuration for the underlying AMQP connection.
	AMQPConfig amqp.Config
}

// DefaultPoolSize is the default size to use for channel pools.
const DefaultPoolSize = 20

// Dial connects to an AMQP-based Overpass network using the default dialer
// and Overpass configuration.
func Dial(dsn string) (overpass.Peer, error) {
	d := Dialer{}
	return d.Dial(context.Background(), dsn, overpass.DefaultConfig)
}

// DialConfig connects to an AMQP-based Overpass network using the default
// dialer and the specified context and Overpass configuration.
func DialConfig(ctx context.Context, dsn string, cfg overpass.Config) (overpass.Peer, error) {
	d := Dialer{}
	return d.Dial(ctx, dsn, cfg)
}

// Dial connects to an AMQP-based Overpass network using d and the specified
// context and Overpass configuration.
func (d *Dialer) Dial(ctx context.Context, dsn string, cfg overpass.Config) (overpass.Peer, error) {
	if dsn == "" {
		dsn = "amqp://localhost"
	}

	amqpCfg := d.AMQPConfig
	if amqpCfg.Properties == nil {
		amqpCfg.Properties = amqp.Table{
			"product": path.Base(os.Args[0]),
			"version": "overpass-go/0.0.0",
		}
	}

	cfg = withDefaults(cfg)

	broker, err := amqp.DialConfig(dsn, amqpCfg)
	if err != nil {
		return nil, err
	}

	defer func() {
		// if an error has occurred when the function exits, close the
		// broker connection immediately, otherwise it is given to the peer
		if err != nil {
			broker.Close()
		}
	}()

	poolSize := d.PoolSize
	if poolSize == 0 {
		poolSize = DefaultPoolSize
	}

	channels := amqputil.NewChannelPool(broker, poolSize)
	peerID, err := d.establishIdentity(ctx, channels, cfg.Logger)
	if err != nil {
		return nil, err
	}

	cfg.Logger.Log(
		"%s connected to '%s' as %s",
		peerID.ShortString(),
		dsn,
		peerID,
	)

	localStore := localsession.NewStore()
	revStore := &revision.AggregateStore{
		PeerID: peerID,
		Local:  localStore,
		// Remote revision store depends on invoker, created below
	}

	invoker, server, err := commandamqp.New(peerID, cfg, revStore, channels)
	if err != nil {
		return nil, err
	}

	notifier, listener, err := notifyamqp.New(peerID, cfg, localStore, revStore, channels)
	if err != nil {
		return nil, err
	}

	remoteStore := remotesession.NewStore(peerID, invoker, cfg.PruneInterval, cfg.Logger)
	revStore.Remote = remoteStore

	remotesession.Listen(server, peerID, localStore, cfg.Logger)

	return newPeer(
		peerID,
		broker,
		localStore,
		remoteStore,
		invoker,
		server,
		notifier,
		listener,
		cfg.Logger,
	), nil
}

// establishIdentity allocates a new peer ID on the broker.
func (d *Dialer) establishIdentity(
	ctx context.Context,
	channels amqputil.ChannelPool,
	logger overpass.Logger,
) (id overpass.PeerID, err error) {
	var channel *amqp.Channel

	for {
		channel, err = channels.Get()
		if err != nil {
			return
		}

		id = overpass.NewPeerID()
		_, err = channel.QueueDeclare(
			id.ShortString(), // this queue is used purely to reserve the peer ID
			false,            // durable
			false,            // autoDelete
			true,             // exclusive,
			false,            // noWait
			nil,              // args
		)

		if amqpErr, ok := err.(*amqp.Error); !ok || amqpErr.Code != amqp.ResourceLocked {
			if err == nil {
				channels.Put(channel)
			}

			return
		}

		select {
		case <-ctx.Done():
			err = ctx.Err()
			return
		default:
			if logger.IsDebug() {
				logger.Log(
					"%s already registered, retrying with a different peer ID",
					id.ShortString(),
				)
			}
		}
	}
}

// withDefaults returns a copy of cfg config with empty properties replaced
// with their defaults.
func withDefaults(cfg overpass.Config) overpass.Config {
	if cfg.DefaultTimeout == 0 {
		cfg.DefaultTimeout = overpass.DefaultConfig.DefaultTimeout
	}

	if cfg.CommandPreFetch == 0 {
		cfg.CommandPreFetch = overpass.DefaultConfig.CommandPreFetch
	}

	if cfg.SessionPreFetch == 0 {
		cfg.SessionPreFetch = overpass.DefaultConfig.SessionPreFetch
	}

	if cfg.Logger == nil {
		cfg.Logger = overpass.DefaultConfig.Logger
	}

	if cfg.PruneInterval == 0 {
		cfg.PruneInterval = overpass.DefaultConfig.PruneInterval
	}

	return cfg
}
