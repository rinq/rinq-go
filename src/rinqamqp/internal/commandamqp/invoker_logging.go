package commandamqp

import (
	"github.com/jmalloc/twelf/src/twelf"
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

func logInvokerInvalidMessageID(
	logger twelf.Logger,
	peerID ident.PeerID,
	msgID string,
) {
	logger.Debug(
		"%s invoker ignored AMQP message, '%s' is not a valid message ID",
		peerID.ShortString(),
		msgID,
	)
}

func logInvokerIgnoredMessage(
	logger twelf.Logger,
	peerID ident.PeerID,
	msgID ident.MessageID,
	err error,
) {
	logger.Debug(
		"%s invoker ignored AMQP message %s, %s",
		peerID.ShortString(),
		msgID.ShortString(),
		err,
	)
}

func logUnicastCallBegin(
	logger twelf.Logger,
	peerID ident.PeerID,
	msgID ident.MessageID,
	target ident.PeerID,
	ns string,
	cmd string,
	traceID string,
	payload *rinq.Payload,
) {
	logger.Debug(
		"%s invoker began unicast '%s::%s' call %s to %s [%s] >>> %s",
		peerID.ShortString(),
		ns,
		cmd,
		msgID.ShortString(),
		target.ShortString(),
		traceID,
		payload,
	)
}

func logBalancedCallBegin(
	logger twelf.Logger,
	peerID ident.PeerID,
	msgID ident.MessageID,
	ns string,
	cmd string,
	traceID string,
	payload *rinq.Payload,
) {
	logger.Debug(
		"%s invoker began '%s::%s' call %s [%s] >>> %s",
		peerID.ShortString(),
		ns,
		cmd,
		msgID.ShortString(),
		traceID,
		payload,
	)
}

func logCallEnd(
	logger twelf.Logger,
	peerID ident.PeerID,
	msgID ident.MessageID,
	ns string,
	cmd string,
	traceID string,
	payload *rinq.Payload,
	err error,
) {
	if !logger.IsDebug() {
		return
	}

	switch e := err.(type) {
	case nil:
		logger.Debug(
			"%s invoker completed '%s::%s' call %s successfully [%s] <<< %s",
			peerID.ShortString(),
			ns,
			cmd,
			msgID.ShortString(),
			traceID,
			payload,
		)
	case rinq.Failure:
		var message string
		if e.Message != "" {
			message = ": " + e.Message
		}

		logger.Debug(
			"%s invoker completed '%s::%s' call %s with '%s' failure%s [%s] <<< %s",
			peerID.ShortString(),
			ns,
			cmd,
			msgID.ShortString(),
			e.Type,
			message,
			traceID,
			payload,
		)
	default:
		logger.Debug(
			"%s invoker completed '%s::%s' call %s with error [%s] <<< %s",
			peerID.ShortString(),
			ns,
			cmd,
			msgID.ShortString(),
			traceID,
			err,
		)
	}
}

func logAsyncRequest(
	logger twelf.Logger,
	peerID ident.PeerID,
	msgID ident.MessageID,
	ns string,
	cmd string,
	traceID string,
	payload *rinq.Payload,
	err error,
) {
	logger.Debug(
		"%s invoker sent asynchronous '%s::%s' call request %s [%s] >>> %s",
		peerID.ShortString(),
		ns,
		cmd,
		msgID.ShortString(),
		traceID,
		payload,
	)
}

func logAsyncResponse(
	logger twelf.Logger,
	peerID ident.PeerID,
	msgID ident.MessageID,
	ns string,
	cmd string,
	traceID string,
	payload *rinq.Payload,
	err error,
) {
	logger.Debug(
		"%s invoker received asynchronous '%s::%s' call response %s [%s] >>> %s",
		peerID.ShortString(),
		ns,
		cmd,
		msgID.ShortString(),
		traceID,
		payload,
	)
}

func logBalancedExecute(
	logger twelf.Logger,
	peerID ident.PeerID,
	msgID ident.MessageID,
	ns string,
	cmd string,
	traceID string,
	payload *rinq.Payload,
	err error,
) {
	logger.Debug(
		"%s invoker sent '%s::%s' execution %s [%s] >>> %s",
		peerID.ShortString(),
		ns,
		cmd,
		msgID.ShortString(),
		traceID,
		payload,
	)
}

func logMulticastExecute(
	logger twelf.Logger,
	peerID ident.PeerID,
	msgID ident.MessageID,
	ns string,
	cmd string,
	traceID string,
	payload *rinq.Payload,
	err error,
) {
	logger.Debug(
		"%s invoker sent multicast '%s::%s' execution %s [%s] >>> %s",
		peerID.ShortString(),
		ns,
		cmd,
		msgID.ShortString(),
		traceID,
		payload,
	)
}

func logInvokerStart(
	logger twelf.Logger,
	peerID ident.PeerID,
	preFetch uint,
) {
	logger.Debug(
		"%s invoker started (pre-fetch: %d)",
		peerID.ShortString(),
		preFetch,
	)
}

func logInvokerStopping(
	logger twelf.Logger,
	peerID ident.PeerID,
	pending int,
) {
	logger.Debug(
		"%s invoker stopping gracefully (pending: %d)",
		peerID.ShortString(),
		pending,
	)
}

func logInvokerStop(
	logger twelf.Logger,
	peerID ident.PeerID,
	err error,
) {
	if err == nil {
		logger.Debug(
			"%s invoker stopped",
			peerID.ShortString(),
		)
	} else {
		logger.Debug(
			"%s invoker stopped: %s",
			peerID.ShortString(),
			err,
		)
	}
}
