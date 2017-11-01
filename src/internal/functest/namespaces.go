package functest

import (
	"fmt"
	"os"
	"sync"

	"github.com/streadway/amqp"
)

var namespaces struct {
	mutex   sync.Mutex
	count   int
	names   map[string]struct{}
	broker  *amqp.Connection
	channel *amqp.Channel
}

// NewNamespace returns a string that is a valid namespace.
func NewNamespace() string {
	namespaces.mutex.Lock()
	defer namespaces.mutex.Unlock()

	namespaces.count++
	ns := fmt.Sprintf("rinq-test-%d-%d", os.Getpid(), namespaces.count)

	if namespaces.names == nil {
		namespaces.names = map[string]struct{}{}
	}

	namespaces.names[ns] = struct{}{}

	return ns
}

// tearDownNamespaces cleans up any queues created for command namespaces made
// via NewNamespace()
func tearDownNamespaces() {
	namespaces.mutex.Lock()
	defer namespaces.mutex.Unlock()

	if len(namespaces.names) == 0 {
		return
	}

	if namespaces.channel == nil {
		dsn := os.Getenv("RINQ_AMQP_DSN")
		if dsn == "" {
			dsn = "amqp://localhost"
		}

		broker, err := amqp.Dial(dsn)
		if err != nil {
			fmt.Println(err)
			return
		}

		channel, err := broker.Channel()
		if err != nil {
			_ = broker.Close()
			fmt.Println(err)
			return
		}

		namespaces.broker = broker
		namespaces.channel = channel
	}

	for ns := range namespaces.names {
		_, err := namespaces.channel.QueueDelete(
			"cmd."+ns, // see commandamqp.balancedRequestQueue()
			false,     // ifUnused,
			false,     // ifEmpty,
			false,     // noWait
		)
		if err != nil {
			namespaces.broker = nil
			namespaces.channel = nil
			fmt.Println(err)
			return
		}

		delete(namespaces.names, ns)
	}
}
