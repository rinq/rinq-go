package main

import "github.com/rinq/rinq-go/src/rinq/amqp"

func main() {
	p, err := amqp.DialEnv()
	if err != nil {
		panic(err)
	}
	defer p.Stop()

	<-p.Done()
}
