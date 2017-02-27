// +build !nobroker

package overpass_test

import (
	"context"
	"fmt"

	. "github.com/over-pass/overpass-go/src/overpass"
	"github.com/over-pass/overpass-go/src/overpass/amqp"
)

// This example illustrates how to read an attribute from a session.
//
// It includes logic necessary to fetch the attribute even if the Revision in
// use is out-of-date, by retrying on the latest revision.
func ExampleRevision_get() {
	peer, err := amqp.Dial("")
	if err != nil {
		panic(err)
	}
	defer peer.Stop()

	sess := peer.Session()
	defer sess.Close()

	rev, err := sess.CurrentRevision()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	var attr Attr
	for {
		attr, err = rev.Get(ctx, "user-id")
		if err != nil {
			if ShouldRetry(err) {
				// the attribute could not be fetched because it has been
				// updated since rev was obtained
				rev, err = rev.Refresh(ctx)
				if err == nil {
					continue
				}
			}
			panic(err)
		}

		break
	}

	if attr.Value == "" {
		fmt.Println("user is not logged in")
	}

	// Output: user is not logged in
}

// This example illustrates how to modify an attribute in a session.
//
// It includes logic to retry in the face of an optimistic-lock failure, which
// occurs if the revision is out of date.
func ExampleRevision_update() {
	peer, err := amqp.Dial("")
	if err != nil {
		panic(err)
	}
	defer peer.Stop()

	sess := peer.Session()
	defer sess.Close()

	rev, err := sess.CurrentRevision()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	for {
		rev, err = rev.Update(ctx, Set("user-id", "123"))
		if err != nil {
			if ShouldRetry(err) {
				// the session could not be updated because rev is out of date
				rev, err = rev.Refresh(ctx)
				if err == nil {
					continue
				}
			}
			panic(err)
		}

		fmt.Printf("updated to revision #%d\n", rev.Ref().Rev)
		break
	}

	// Output: updated to revision #1
}
