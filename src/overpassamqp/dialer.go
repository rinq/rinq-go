package overpassamqp

import (
	"context"
	"log"

	"github.com/over-pass/overpass-go/src/internals"
	"github.com/over-pass/overpass-go/src/internals/amqputil"
	"github.com/over-pass/overpass-go/src/internals/command"
	"github.com/over-pass/overpass-go/src/internals/notify"
	"github.com/over-pass/overpass-go/src/overpass"
	"github.com/streadway/amqp"
)

// Dialer creates a peer by connecting to an AMQP broker.
type Dialer struct {
	// The number of AMQP channels to keep open when not in use.
	PoolSize uint

	// Low-level AMQP configuration.
	AMQPConfig amqp.Config
}

// Dial connects to an Overpass network and returns a new peer using the default dialer.
func Dial(ctx context.Context, dsn string, config overpass.Config) (overpass.Peer, error) {
	d := Dialer{}
	return d.Dial(ctx, dsn, config)
}

// Dial connects to an Overpass network and returns a new peer.
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
	// in the AMQP context.
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

	localStore := &localStore{}
	remoteStore := &remoteStore{}
	revStore := internals.NewAggregateRevisionStore(peerID, localStore, remoteStore)

	invoker, server, err := command.New(peerID, config, revStore, channels)
	if err != nil {
		return nil, err
	}

	notifier, listener, err := notify.New(peerID, config, localStore, revStore, channels)
	if err != nil {
		return nil, err
	}

	config.Logger.Printf(
		"%s peer connected to '%s' as %s",
		peerID.ShortString(),
		dsn,
		peerID,
	)

	return newPeer(
		peerID,
		broker,
		localStore,
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
	logger *log.Logger,
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
			logger.Printf(
				"%s peer already registered, retrying",
				id.ShortString(),
			)
		}
	}
}
