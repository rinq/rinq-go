package notifyamqp

const (
	// unicastExchange is the exchange used to publish notifications directly to
	// a specific session.
	unicastExchange = "ntf.uc"

	// multicastExchange is the exchange used to publish notifications that are
	// sent to multiple sessions based on a rinq.Constraint.
	multicastExchange = "ntf.mc"
)
