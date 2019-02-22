package pubsub

import (
	"container/list"
	"encoding/json"
	"sync"
)

type client struct {
	m              sync.Mutex
	clientMessages map[topicName]*list.List
}

func newClient() *client {
	return &client{
		clientMessages: map[topicName]*list.List{},
	}
}

func (c *client) subscribe(topic topicName) {
	c.m.Lock()
	queue := c.clientMessages[topic]
	// To avoid queue rewrite if client subscribes two times on same subscription
	if queue == nil {
		c.clientMessages[topic] = list.New()
	}
	c.m.Unlock()
}

// unsubscribe returns amount of topics left after unsubscribtion
func (c *client) unsubscribe(topic topicName) int {
	c.m.Lock()
	delete(c.clientMessages, topic)
	topicsAmount := len(c.clientMessages)
	c.m.Unlock()
	return topicsAmount
}

func (c *client) publish(topic topicName, msg json.RawMessage) {
	c.m.Lock()
	queue := c.clientMessages[topic]
	queue.PushBack(msg)
	c.m.Unlock()
}

func (c *client) poll(topic topicName) (json.RawMessage, error) {
	c.m.Lock()
	defer c.m.Unlock()
	queue := c.clientMessages[topic]
	if queue == nil {
		return nil, ErrSubscriptionNotFound
	}
	e := queue.Front()
	if e == nil {
		return nil, nil
	}
	msg := e.Value.(json.RawMessage)
	queue.Remove(e)
	return msg, nil
}
