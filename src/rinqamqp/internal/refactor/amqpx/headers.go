package amqpx

import (
	"fmt"

	"github.com/streadway/amqp"
)

// SetHeader sets a message header with the given key and value, initializing
// the message headers if they are nil.
func SetHeader(msg *amqp.Publishing, key string, value interface{}) {
	if msg.Headers == nil {
		msg.Headers = amqp.Table{}
	}

	msg.Headers[key] = value
}

// GetHeaderString returns the header with the given key, asserting that it is
// a string.
func GetHeaderString(msg *amqp.Delivery, key string) (string, error) {
	v, ok := msg.Headers[key]
	if !ok {
		return "", fmt.Errorf("%s header is not present", key)
	}

	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("%s header is not a string", key)
	}

	return s, nil
}

// GetHeaderBytes returns the header with the given key, asserting that it is
// a byte slice.
func GetHeaderBytes(msg *amqp.Delivery, key string) ([]byte, error) {
	v, ok := msg.Headers[key]
	if !ok {
		return nil, fmt.Errorf("%s header is not present", key)
	}

	b, ok := v.([]byte)
	if !ok {
		return nil, fmt.Errorf("%s header is not a byte slice", key)
	}

	return b, nil
}

// GetHeaderBytesOptional returns the header with the given key, asserting that
// it is a byte slice.
func GetHeaderBytesOptional(msg *amqp.Delivery, key string) ([]byte, bool, error) {
	v, ok := msg.Headers[key]
	if !ok {
		return nil, false, nil
	}

	b, ok := v.([]byte)
	if !ok {
		return nil, false, fmt.Errorf("%s header is not a byte slice", key)
	}

	return b, true, nil
}
