package notifications

// Subscriber is an implementation of transport.Subscriber that receives
// notifications from the AMQP broker.
type Subscriber struct {
	Binder
	Consumer
}
