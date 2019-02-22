package pubsub

import (
	"fmt"
	"log"
)

func ExampleServer() {
	s := NewServer()
	topic := "topic"
	subscriber := "sub"
	msg := []byte(`{"foo":"bar"}`)
	s.Subscribe(topic, subscriber)
	s.Publish(topic, msg)
	msg, err := s.Poll(topic, subscriber)
	if err != nil {
		log.Fatalf("Unexpected error: %v", err)
	}
	s.Unsubscribe(topic, subscriber)

	fmt.Println(string(msg))
	// Output: {"foo":"bar"}
}
