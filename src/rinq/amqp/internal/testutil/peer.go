package testutil

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

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

var (
	mutex   sync.Mutex
	broker  *amqplib.Connection
	nsCount int
	nsClean int
)

// Namespace returns a unique namespace name.
func Namespace() string {
	mutex.Lock()
	defer mutex.Unlock()

	nsCount++

	return namespace(nsCount)
}

func namespace(i int) string {
	return fmt.Sprintf("rinq-test-%d-%d", os.Getpid(), i)
}

// TearDown cleans up any AMQP resources that should not persist between tests.
func TearDown() {
	mutex.Lock()
	defer mutex.Unlock()

	if broker == nil {
		dsn := os.Getenv("RINQ_AMQP_DSN")
		if dsn == "" {
			dsn = "amqp://localhost"
		}

		var err error
		broker, err = amqplib.Dial(dsn)
		if err != nil {
			broker = nil
			fmt.Println(err)
			return
		}
	}

	channel, err := broker.Channel()
	defer channel.Close()

	if err != nil {
		broker = nil
		fmt.Println(err)
		return
	}

	for nsClean <= nsCount {
		_, err = channel.QueueDelete(
			"cmd."+namespace(nsClean), // see commandamqp.balancedRequestQueue()
			false, // ifUnused,
			false, // ifEmpty,
			false, // noWait
		)
		if err != nil {
			broker = nil
			fmt.Println(err)
			return
		}
		nsClean++
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
