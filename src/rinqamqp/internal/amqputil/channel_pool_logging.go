package amqputil

import (
	"github.com/rinq/rinq-go/src/rinq"
)

func logChannelPoolGet(
	logger rinq.Logger,
	remaining int,
	err error,
) {
	if !logger.IsDebug() {
		return
	}

	if err != nil {
		logger.Log(
			"channel pool get (remaining: %d): %s",
			remaining,
			err,
		)
	} else {
		logger.Log(
			"channel pool get (remaining: %d)",
			remaining,
		)
	}
}

func logChannelPoolGetQOS(
	logger rinq.Logger,
	remaining int,
	err error,
) {
	if !logger.IsDebug() {
		return
	}

	if err != nil {
		logger.Log(
			"channel pool get QOS (remaining: %d): %s",
			remaining,
			err,
		)
	} else {
		logger.Log(
			"channel pool get QOS (remaining: %d)",
			remaining,
		)
	}
}

func logChannelPoolPut(
	logger rinq.Logger,
	remaining int,
	err error,
) {
	if !logger.IsDebug() {
		return
	}

	if err != nil {
		logger.Log(
			"channel pool put (remaining: %d): %s",
			remaining,
			err,
		)
	} else {
		logger.Log(
			"channel pool put (remaining: %d)",
			remaining,
		)
	}
}

func logChannelPoolCleanup(
	logger rinq.Logger,
	remaining int,
	err error,
) {
	if !logger.IsDebug() {
		return
	}

	if err != nil {
		logger.Log(
			"channel pool cleanup (remaining: %d): %s",
			remaining,
			err,
		)
	} else {
		logger.Log(
			"channel pool cleanup (remaining: %d)",
			remaining,
		)
	}
}

func logChannelPoolStart(
	logger rinq.Logger,
	size int,
) {
	if !logger.IsDebug() {
		return
	}

	logger.Log(
		"channel pool started (size: %d)",
		size,
	)
}

func logChannelPoolGraceful(
	logger rinq.Logger,
	remaining int,
) {
	if !logger.IsDebug() {
		return
	}

	logger.Log(
		"channel pool stopped gracefully (remaining: %d)",
		remaining,
	)
}

func logChannelPoolStop(
	logger rinq.Logger,
	remaining int,
	err error,
) {
	if !logger.IsDebug() {
		return
	}

	if err == nil {
		logger.Log(
			"channel pool stopped (remaining: %d)",
			remaining,
		)
	} else {
		logger.Log(
			"channel pool stopped (remaining: %d): %s",
			remaining,
			err,
		)
	}
}
