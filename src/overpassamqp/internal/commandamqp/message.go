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
