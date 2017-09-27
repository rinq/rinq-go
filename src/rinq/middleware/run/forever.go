package run

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rinq/rinq-go/src/rinq"
)

const (
	// DefaultRetryDelay is the default time to wait between dialing.
	DefaultRetryDelay = 5 * time.Second

	// DefaultStopTimeout is the default time to wait for a peer to shutdown
	// gracefully.
	DefaultStopTimeout = 15 * time.Second
)

// Dialer is a function that establishes a peer on a Rinq network.
type Dialer func(context.Context) (rinq.Peer, error)

// Forever repeatedly establishes a new peer using dialer until a signal
// is received or the peer is shutdown cleanly.
func Forever(
	ctx context.Context,
	dial Dialer,
	t time.Duration,
	d time.Duration,
) error {
	f := forever{
		parent:      ctx,
		dialer:      dial,
		stopTimeout: t,
		retryDelay:  d,
	}

	return f.run()
}

type state func() (state, error)

type forever struct {
	parent      context.Context
	dialer      Dialer
	stopTimeout time.Duration
	retryDelay  time.Duration
	firstRetry  bool
	peer        rinq.Peer
	signals     chan os.Signal
}

func (f *forever) run() (err error) {
	f.signals = make(chan os.Signal, 2)
	defer signal.Stop(f.signals)

	signal.Notify(f.signals, os.Interrupt)
	signal.Notify(f.signals, syscall.SIGTERM)
	signal.Notify(f.signals, syscall.SIGQUIT)

	s := f.dial

	for s != nil && err == nil {
		s, err = s()
	}

	return
}

func (f *forever) dial() (state, error) {
	ctx, cancel := context.WithCancel(f.parent)
	defer cancel()

	var (
		err  error
		done = make(chan struct{})
	)

	go func() {
		f.peer, err = f.dialer(ctx)
		close(done)
	}()

	select {
	case <-f.signals:
		cancel()
		return nil, nil

	case <-done:
		if err == nil {
			return f.wait, nil
		}

		return f.retry, nil
	}
}

func (f *forever) wait() (state, error) {
	f.firstRetry = true

	select {
	case sig := <-f.signals:
		if sig == syscall.SIGQUIT {
			return f.forceful, nil
		}

		return f.graceful, nil

	case <-f.parent.Done():
		return f.forceful, nil

	case <-f.peer.Done():
		if err := f.peer.Err(); err != nil {
			return f.retry, nil
		}

		return nil, nil
	}
}

func (f *forever) graceful() (state, error) {
	f.peer.GracefulStop()

	t := f.stopTimeout
	if t == 0 {
		t = DefaultStopTimeout
	}

	select {
	case <-time.After(t):
		return f.forceful, nil

	case <-f.signals:
		return f.forceful, nil

	case <-f.parent.Done():
		return f.forceful, nil

	case <-f.peer.Done():
		return nil, f.peer.Err()
	}
}

func (f *forever) forceful() (state, error) {
	f.peer.Stop()
	<-f.peer.Done()

	return nil, f.peer.Err()
}

func (f *forever) retry() (state, error) {
	if f.firstRetry {
		f.firstRetry = false
		return f.dial, nil
	}

	t := f.retryDelay
	if t == 0 {
		t = DefaultRetryDelay
	}

	select {
	case <-time.After(t):
		return f.dial, nil

	case <-f.signals:
		return nil, nil

	case <-f.parent.Done():
		return nil, nil
	}
}
