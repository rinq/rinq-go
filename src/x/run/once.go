package run

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rinq/rinq-go/src/rinq"
)

// Once blocks until p is done, or a signal is received.
//
// If a SIGINT or SIGTERM is received, p is stopped gracefully. If a SIGQUIT is
// received p is stopped forcefully.
//
// If a second SIGINT, SIGTERM or SIGQUIT is received; or the stop timeout t is
// reached while the peer is stopping gracefully, the peer is stopped forcefully.
//
// If t is 0, DefaultStopTimeout is used.
func Once(p rinq.Peer, t time.Duration) error {
	if t == 0 {
		t = DefaultStopTimeout
	}

	sig := make(chan os.Signal, 5)
	defer signal.Stop(sig)
	signal.Notify(sig, os.Interrupt)
	signal.Notify(sig, syscall.SIGTERM)
	signal.Notify(sig, syscall.SIGQUIT)

	select {
	case s := <-sig: // received first signal
		if s != syscall.SIGQUIT {
			p.GracefulStop()

			select {
			case <-sig: // second signal received
			case <-time.After(t): // stop timeout expired
			case <-p.Done(): // the graceful stop has (likely) completed
				return p.Err()
			}
		}

		p.Stop()
		<-p.Done()
		return p.Err()

	case <-p.Done(): // the peer completed before any signal was received
		return p.Err()
	}
}
