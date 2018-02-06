package amqpx

import "github.com/streadway/amqp"

// ChannelPool provides a pool of reusable AMQP channels.
type ChannelPool interface {
	// Get fetches a channel from the pool, or creates one as necessary.
	Get() (*amqp.Channel, error)

	// Put returns a channel to the pool.
	Put(*amqp.Channel)
}
