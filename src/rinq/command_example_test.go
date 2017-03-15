// +build !without_amqp,!without_examples

package rinq_test

import (
	"context"
	"fmt"

	. "github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/amqp"
)

// This example illustrates how to respond to a command request with an
// application-defined failure.
func ExampleResponse_fail() {
	peer, err := amqp.Dial("")
	if err != nil {
		panic(err)
	}
	defer peer.Stop()

	peer.Listen("my-api", func(
		ctx context.Context,
		req Request,
		res Response,
	) {
		defer req.Payload.Close()

		res.Fail(
			"my-api-error",
			"the call to %s failed spectacularly!",
			req.Command,
		)
	})

	sess := peer.Session()
	defer sess.Destroy()

	in, err := sess.Call(context.Background(), "my-api", "test", nil)
	defer in.Close()

	fmt.Println(err)
	// Output: my-api-error: the call to test failed spectacularly!
}
