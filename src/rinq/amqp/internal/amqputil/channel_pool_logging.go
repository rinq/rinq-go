package amqputil

import (
	"errors"
)

func logChannelPoolStart(
	logger rinq.Logger,
) {
	if !logger.IsDebug() {
		return
	}

	logger.Log(
		"channel pool started",
	)
}

func logChannelPoolStop(
	logger rinq.Logger,
	err error,
) {
	if !logger.IsDebug() {
		return
	}

	if err == nil {
		logger.Log("channel pool stopped")
	} else {
		logger.Log(
			"channel pool stopped: %s",
			err,
		)
	}
}
