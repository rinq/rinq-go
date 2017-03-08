package commandamqp

import (
	"context"

	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/trace"
)

func logServerInvalidMessageID(
	logger rinq.Logger,
	peerID ident.PeerID,
	msgID string,
) {
	if !logger.IsDebug() {
		return
	}

	logger.Log(
		"%s server ignored AMQP message, '%s' is not a valid message ID",
		peerID.ShortString(),
		msgID,
	)
}

func logIgnoredMessage(
	logger rinq.Logger,
	peerID ident.PeerID,
	msgID ident.MessageID,
	err error,
) {
	if !logger.IsDebug() {
		return
	}

	logger.Log(
		"%s server ignored AMQP message %s, %s",
		peerID.ShortString(),
		msgID.ShortString(),
		err,
	)
}

func logRequestBegin(
	ctx context.Context,
	logger rinq.Logger,
	peerID ident.PeerID,
	msgID ident.MessageID,
	req rinq.Request,
) {
	if !logger.IsDebug() {
		return
	}

	logger.Log(
		"%s server began '%s::%s' command request %s [%s] <<< %s",
		peerID.ShortString(),
		req.Namespace,
		req.Command,
		msgID.ShortString(),
		trace.Get(ctx),
		req.Payload,
	)
}

func logRequestEnd(
	ctx context.Context,
	logger rinq.Logger,
	peerID ident.PeerID,
	msgID ident.MessageID,
	req rinq.Request,
	payload *rinq.Payload,
	err error,
) {
	if !logger.IsDebug() {
		return
	}

	switch e := err.(type) {
	case nil:
		logger.Log(
			"%s server completed '%s::%s' command request %s successfully [%s] >>> %s",
			peerID.ShortString(),
			req.Namespace,
			req.Command,
			msgID.ShortString(),
			trace.Get(ctx),
			payload,
		)
	case rinq.Failure:
		var message string
		if e.Message != "" {
			message = ": " + e.Message
		}

		logger.Log(
			"%s server completed '%s::%s' command request %s with '%s' failure%s [%s] <<< %s",
			peerID.ShortString(),
			req.Namespace,
			req.Command,
			msgID.ShortString(),
			e.Type,
			message,
			trace.Get(ctx),
			payload,
		)
	default:
		logger.Log(
			"%s server completed '%s::%s' command request %s with error [%s] <<< %s",
			peerID.ShortString(),
			req.Namespace,
			req.Command,
			msgID.ShortString(),
			trace.Get(ctx),
			err,
		)
	}
}

func logNoLongerListening(
	logger rinq.Logger,
	peerID ident.PeerID,
	msgID ident.MessageID,
	ns string,
) {
	if !logger.IsDebug() {
		return
	}

	logger.Log(
		"%s is no longer listening to '%s' namespace, request %s has been re-queued",
		peerID.ShortString(),
		ns,
		msgID.ShortString(),
	)
}

func logRequestRequeued(
	ctx context.Context,
	logger rinq.Logger,
	peerID ident.PeerID,
	msgID ident.MessageID,
	req rinq.Request,
) {
	if !logger.IsDebug() {
		return
	}

	logger.Log(
		"%s did not write a response for '%s::%s' command request, request %s has been re-queued [%s]",
		peerID.ShortString(),
		req.Namespace,
		req.Command,
		msgID.ShortString(),
		trace.Get(ctx),
	)
}

func logRequestRejected(
	ctx context.Context,
	logger rinq.Logger,
	peerID ident.PeerID,
	msgID ident.MessageID,
	req rinq.Request,
	reason string,
) {
	logger.Log(
		"%s did not write a response for '%s::%s' command request %s, request has been abandoned (%s) [%s]",
		peerID.ShortString(),
		req.Namespace,
		req.Command,
		msgID.ShortString(),
		reason,
		trace.Get(ctx),
	)
}

func logServerStart(
	logger rinq.Logger,
	peerID ident.PeerID,
	preFetch int,
) {
	if !logger.IsDebug() {
		return
	}

	logger.Log(
		"%s server started with (pre-fetch: %d)",
		peerID.ShortString(),
		preFetch,
	)
}

func logServerStopping(
	logger rinq.Logger,
	peerID ident.PeerID,
	pending uint,
) {
	if !logger.IsDebug() {
		return
	}

	logger.Log(
		"%s server is stopping gracefully (pending: %d)",
		peerID.ShortString(),
		pending,
	)
}

func logServerStop(
	logger rinq.Logger,
	peerID ident.PeerID,
	err error,
) {
	if !logger.IsDebug() {
		return
	}

	if err == nil {
		logger.Log(
			"%s server stopped",
			peerID.ShortString(),
		)
	} else {
		logger.Log(
			"%s server stopped: %s",
			peerID.ShortString(),
			err,
		)
	}
}
