package testutil

import (
	"context"
	"os"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/amqp"
	amqplib "github.com/streadway/amqp"
)

// NewPeer returns the peer to use for executing functional tests.
func NewPeer() rinq.Peer {
	peer, err := amqp.DialConfig(
		context.Background(),
		os.Getenv("RINQ_AMQP_DSN"),
		rinq.Config{
			Logger: rinq.NewLogger(true),
		},
	)

	if err != nil {
		panic(err)
	}

	return peer
}

// TearDown cleans up any AMQP resources that should not persist between tests.
func TearDown() {
	dsn := os.Getenv("RINQ_AMQP_DSN")
	if dsn == "" {
		dsn = "amqp://localhost"
	}

	broker, err := amqplib.Dial(dsn)
	if err != nil {
		return
	}

	channel, err := broker.Channel()
	if err != nil {
		return
	}

	// see commandamqp.balancedRequestQueue()
	_, _ = channel.QueueDelete(
		"cmd.rinq-func-test",
		false, // ifUnused,
		false, // ifEmpty,
		false, // noWait
	)
}
