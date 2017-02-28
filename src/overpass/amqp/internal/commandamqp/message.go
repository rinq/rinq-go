package commandamqp

const (
	// successResponse is the AMQP message type used for successful call
	// responses.
	successResponse = "s"

	// successResponse is the AMQP message type used for call responses
	// indicating failure for an "expected" application-defined reason.
	failureResponse = "f"

	// successResponse is the AMQP message type used for call responses
	// indicating unepected error or internal error.
	errorResponse = "e"
)

const (
	// namespaceHeader specifies the namespace in unicast command requests.
	namespaceHeader = "n"

	// failureTypeHeader specifies the failure type in command responses with
	// the "failureResponse" type.
	failureTypeHeader = "t"

	// failureMessageHeader holds the error message in command responses with
	// the "failureResponse" type.
	failureMessageHeader = "m"
)

type replyType string

const (
	// replyNone is the AMQP reply-to value used for command requests that are
	// not expecting a reply.
	replyNone replyType = ""

	// replyNormal is the AMQP reply-to value used for command requests that are
	// waiting for a reply.
	replyCorrelated replyType = "c"

	// replyUncorrelated is the AMQP reply-to value used for command requests
	// that are waiting for a reply, but where the invoker does not have
	// any information about the request. This instruct the server to include
	// request information in the response.
	replyUncorrelated replyType = "u"
)
