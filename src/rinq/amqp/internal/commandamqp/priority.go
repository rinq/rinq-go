package commandamqp

const (
	// executePriority is the AMQP priority for "Execute*" operations.
	executePriority uint8 = iota

	// callBalancedPriority is the AMQP priority for "CallBalanced" operations.
	// These operations always have a timeout, so the priority is raised above
	// operations that don't.
	callBalancedPriority

	// callUnicastPriority is the AMQP priority for "CallUnicast" operations.
	// Like CallBalanced, these operations have a timeout. They are also used
	// to implement internal features, and so the priority is raised even higher
	// again.
	callUnicastPriority

	// priorityCount is the number of priorities in use, used to declare the
	// AMQP queues with the exact number of priority slots.
	priorityCount
)
