package amqp

import (
	"context"
	"fmt"
	"net"
	"os"
	"path"
	"time"

	version "github.com/hashicorp/go-version"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/amqp/internal/amqputil"
	"github.com/rinq/rinq-go/src/rinq/amqp/internal/commandamqp"
	"github.com/rinq/rinq-go/src/rinq/amqp/internal/notifyamqp"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/internal/envutil"
	"github.com/rinq/rinq-go/src/rinq/internal/localsession"
	"github.com/rinq/rinq-go/src/rinq/internal/optutil"
	"github.com/rinq/rinq-go/src/rinq/internal/remotesession"
	"github.com/rinq/rinq-go/src/rinq/internal/revision"
	"github.com/rinq/rinq-go/src/rinq/options"
	"github.com/streadway/amqp"
)

// Dialer connects to an AMQP-based Rinq network, establishing the peer's unique
// identity on the network.
type Dialer struct {
	// The minimum number of AMQP channels to keep open. If PoolSize is zero,
	// DefaultPoolSize is used.
	PoolSize uint

	// Configuration for the underlying AMQP connection.
	AMQPConfig amqp.Config
}

const (
	// DefaultDSN is the AMQP DSN to use when no other DSN is specified.
	DefaultDSN = "amqp://localhost"

	// DefaultPoolSize is the default size to use for channel pools.
	DefaultPoolSize = 20
)

// Dial connects to an AMQP-based Rinq network using the default dialer.
func Dial(dsn string, opts ...options.Option) (rinq.Peer, error) {
	d := Dialer{}
	return d.Dial(context.Background(), dsn, opts...)
}

// DialEnv connects to an AMQP-based Rinq network using the a dialer and
// peer options described by environment variables.
//
// The AMQP-specific environment variables are listed below. If any variable is
// undefined, the default value is used. Additionally, Rinq peer options are
// obtained by calling options.FromEnv().
//
// - RINQ_AMQP_DSN
// - RINQ_AMQP_HEARTBEAT (duration in milliseconds, non-zero)
// - RINQ_AMQP_CHANNELS (channel pool size, positive integer, non-zero)
// - RINQ_AMQP_CONNECTION_TIMEOUT (duration in milliseconds, non-zero)
//
// Note that for consistency with other environment variables, RINQ_AMQP_HEARTBEAT
// is specified in milliseconds, but AMQP only supports 1-second resolution for
// heartbeats. The heartbeat value is ROUNDED UP to the nearest whole second.
//
// Options defined by environment variables take precedence over those in the
// opts slice.
func DialEnv(opts ...options.Option) (rinq.Peer, error) {
	d := Dialer{}

	hb, ok, err := envutil.Duration("RINQ_AMQP_HEARTBEAT")
	if err != nil {
		return nil, err
	} else if ok {
		d.AMQPConfig.Heartbeat = hb

		// round up to the nearest second
		if r := d.AMQPConfig.Heartbeat % time.Second; r != 0 {
			d.AMQPConfig.Heartbeat += time.Second - r
		}
	}

	chans, ok, err := envutil.UInt("RINQ_AMQP_CHANNELS")
	if err != nil {
		return nil, err
	} else if ok {
		d.PoolSize = chans
	}

	ctx := context.Background()

	timeout, ok, err := envutil.Duration("RINQ_AMQP_CONNECTION_TIMEOUT")
	if err != nil {
		return nil, err
	} else if ok {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	envOpts, err := options.FromEnv()
	if err != nil {
		return nil, err
	}

	return d.Dial(
		ctx,
		os.Getenv("RINQ_AMQP_DSN"),
		append(opts, envOpts...)...,
	)
}

// Dial connects to an AMQP-based Rinq network using the specified context and
// configuration.
func (d *Dialer) Dial(
	ctx context.Context,
	dsn string,
	opts ...options.Option,
) (rinq.Peer, error) {
	if dsn == "" {
		dsn = DefaultDSN
	}

	cfg, err := optutil.NewConfig(opts...)
	if err != nil {
		return nil, err
	}

	amqpCfg := d.AMQPConfig
	if amqpCfg.Properties == nil {
		product := cfg.Product
		if product == "" {
			product = path.Base(os.Args[0])
		}

		amqpCfg.Properties = amqp.Table{
			"product": product,
			"version": "rinq-go/" + rinq.Version,
		}
	}

	if amqpCfg.Dial == nil {
		amqpCfg.Dial = makeDeadlineDialer(ctx)
	}

	broker, err := amqp.DialConfig(dsn, amqpCfg)
	if err != nil {
		return nil, err
	}

	defer func() {
		// if an error has occurred when the function exits, close the
		// broker connection immediately, otherwise it is given to the peer
		if err != nil {
			_ = broker.Close()
		}
	}()

	if err = d.checkCapabilities(broker); err != nil {
		return nil, err
	}

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

	invoker, server, err := commandamqp.New(peerID, cfg, localStore, revStore, channels)
	if err != nil {
		return nil, err
	}

	notifier, listener, err := notifyamqp.New(peerID, cfg, localStore, revStore, channels)
	if err != nil {
		return nil, err
	}

	remoteStore := remotesession.NewStore(peerID, invoker, cfg.PruneInterval, cfg.Logger)
	revStore.Remote = remoteStore

	if err := remotesession.Listen(server, peerID, localStore, cfg.Logger); err != nil {
		return nil, err
	}

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
	logger rinq.Logger,
) (id ident.PeerID, err error) {
	var channel *amqp.Channel

	for {
		channel, err = channels.Get()
		if err != nil {
			return
		}

		id = ident.NewPeerID()
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

func (d *Dialer) checkCapabilities(broker *amqp.Connection) error {
	product, _ := broker.Properties["product"].(string)

	ver, _ := broker.Properties["version"].(string)
	semver, err := version.NewVersion(ver)
	if err != nil {
		return err
	}

	var minVersion *version.Version

	switch product {
	case "RabbitMQ":
		// minimum of 3.5.0 is required for priority queues
		minVersion = version.Must(version.NewVersion("3.5.0"))
	default:
		return fmt.Errorf("unsupported AMQP broker: %s", product)
	}

	if semver.LessThan(minVersion) {
		return fmt.Errorf(
			"unsupported AMQP broker: %s %s, minimum version is %s",
			product,
			semver.String(),
			minVersion.String(),
		)
	}

	return nil
}

type amqpDialer func(network, addr string) (net.Conn, error)

// makeDeadlineDialer returns a dial function suitable for use in amqp.Config.Dial
// which honors the deadline in ctx.
func makeDeadlineDialer(ctx context.Context) amqpDialer {
	dl, ok := ctx.Deadline()
	if !ok {
		// if there is no deadline, return nil, thereby using the default
		// dialer provided by the amqp package.
		return nil
	}

	return func(network, addr string) (conn net.Conn, err error) {
		d := net.Dialer{}
		conn, err = d.DialContext(ctx, network, addr)

		if err == nil {
			err = conn.SetDeadline(dl)
		}

		return
	}
}
