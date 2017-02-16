package commandamqp

const (
	// executePriority is the AMQP priority for "Execute*" operations.
	executePriority = iota

	// callPriority is the AMQP priority for "Call*" operations. These
	// operations always have a timeout, so the priority is raised above
	// operations that don't.
	callPriority

	// priorityCount is the number of priorities in use, used to declare the
	// AMQP queues with the exact number of priority slots. // TODO
	priorityCount
)
