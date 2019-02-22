package pubsub

import (
	"sync"
)

// Server should be created via pubsub.NewServer fuction
type Server struct {
	m       sync.RWMutex
	clients map[subscriberName]*client
	topics  map[topicName]map[subscriberName]*client
}

// Subscribe - subscribes subscriber to topic.
// Subscriber can receive messages received after subscription on this topic with Poll
func (s *Server) Subscribe(topic string, subscriber string) {
	tn, sn := topicSubscriberNames(topic, subscriber)
	s.m.Lock()
	cl := s.clients[sn]
	if cl == nil {
		cl = s.clients[sn]
		if cl == nil {
			cl = newClient()
			s.clients[sn] = cl
		}
	}
	cl.subscribe(tn)
	clients := s.topics[tn]
	if clients == nil {
		clients = map[subscriberName]*client{}
		s.topics[tn] = clients
	}
	clients[sn] = cl
	s.m.Unlock()
}

// Unsubscribe - remove topic's subscription for client
func (s *Server) Unsubscribe(topic, subscriber string) {
	tn, sn := topicSubscriberNames(topic, subscriber)
	s.m.Lock()
	cl := s.clients[sn]
	topicsAmount := cl.unsubscribe(tn)
	if topicsAmount == 0 {
		delete(s.clients, sn)
	}
	topicClients := s.topics[tn]
	delete(topicClients, sn)
	if len(topicClients) == 0 {
		delete(s.topics, tn)
	}
	s.m.Unlock()
}

// Publish - deliver msg to all topic subscribers
// This function traverse all topic subscribers,
// so it is better to call it in async manner: go s.Publish(topic, msg)
func (s *Server) Publish(topic string, msg []byte) {
	tn := topicName(topic)
	s.m.RLock()
	clients := s.topics[tn]
	for _, cl := range clients {
		cl.publish(tn, msg)
	}
	s.m.RUnlock()
}

// Poll - returns next message on topic of subscriber and removes it from subscribers queue.
// If there are no new messages it will return nil, nil.
// If there is no subscription it will return pubsub.ErrSubscriptionNotFound
func (s *Server) Poll(topic, subscriber string) ([]byte, error) {
	tn, sn := topicSubscriberNames(topic, subscriber)
	s.m.RLock()
	cl := s.clients[sn]
	s.m.RUnlock()
	if cl == nil {
		return nil, ErrSubscriptionNotFound
	}
	return cl.poll(tn)
}
