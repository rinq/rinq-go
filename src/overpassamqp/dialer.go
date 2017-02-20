package overpassamqp

import (
	"context"

	"github.com/over-pass/overpass-go/src/internals/amqputil"
	"github.com/over-pass/overpass-go/src/internals/localsession"
	"github.com/over-pass/overpass-go/src/internals/remotesession"
	"github.com/over-pass/overpass-go/src/internals/revision"
	"github.com/over-pass/overpass-go/src/overpass"
	"github.com/over-pass/overpass-go/src/overpassamqp/commandamqp"
	"github.com/over-pass/overpass-go/src/overpassamqp/notifyamqp"
	"github.com/streadway/amqp"
)

// Dialer creates a peer by connecting to an AMQP broker.
type Dialer struct {
	// The number of AMQP channels to keep open when not in use.
	PoolSize uint

	// Low-level AMQP configuration.
	AMQPConfig amqp.Config
}

// Dial connects to an AMQP broker and returns a new peer using the default dialer.
func Dial(ctx context.Context, dsn string, config overpass.Config) (overpass.Peer, error) {
	d := Dialer{}
	return d.Dial(ctx, dsn, config)
}

// Dial connects to an AMQP broker and returns a new peer.
func (d *Dialer) Dial(ctx context.Context, dsn string, config overpass.Config) (overpass.Peer, error) {
	if dsn == "" {
		dsn = "amqp://localhost"
	}

	amqpConfig := d.AMQPConfig
	if amqpConfig.Properties == nil {
		amqpConfig.Properties = amqp.Table{
			"product": "overpass",
			"version": "golang/2.0.0",
		}
	}

	config = config.WithDefaults()

	// TODO: honour ctx deadline here, possibly by provided a custom Dial func
	// in the AMQP config.
	broker, err := amqp.DialConfig(dsn, amqpConfig)
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
		poolSize = 20
	}

	channels := amqputil.NewChannelPool(broker, poolSize)
	peerID, err := d.establishIdentity(ctx, channels, config.Logger)
	if err != nil {
		return nil, err
	}

	sessions := localsession.NewStore()
	revisions := &revision.AggregateStore{
		PeerID: peerID,
		Local:  sessions,
		// Remote revision store depends on invoker, created below
	}

	invoker, server, err := commandamqp.New(peerID, config, revisions, channels)
	if err != nil {
		return nil, err
	}

	notifier, listener, err := notifyamqp.New(peerID, config, sessions, revisions, channels)
	if err != nil {
		return nil, err
	}

	revisions.Remote = remotesession.NewStore(peerID, invoker)

	remotesession.Listen(peerID, sessions, server)

	config.Logger.Log(
		"%s peer connected to '%s' as %s",
		peerID.ShortString(),
		dsn,
		peerID,
	)

	return newPeer(
		peerID,
		broker,
		sessions,
		invoker,
		server,
		notifier,
		listener,
		config.Logger,
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
					"%s peer already registered, retrying",
					id.ShortString(),
				)
			}
		}
	}
}
