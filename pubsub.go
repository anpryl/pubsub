package pubsub

import (
	"errors"
)

var ErrSubscriptionNotFound = errors.New("pubsub: subscription not found")

// NewServer initialize pubsub server. Please create Server only via NewServer
func NewServer() *Server {
	return &Server{
		clients: map[subscriberName]*client{},
		topics:  map[topicName]map[subscriberName]*client{},
	}
}

// subscriberName - type alias not to get confused between topic string and subscriber string
type subscriberName string

// topicName - type alias not to get confused between topic string and subscriber string
type topicName string

// topicSubscriberNames - takes topic and subscriber strigns and returns separate types
func topicSubscriberNames(topic, subscriber string) (topicName, subscriberName) {
	return topicName(topic), subscriberName(subscriber)
}
