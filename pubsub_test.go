package pubsub

import (
	"bytes"
	"fmt"
	"sync"
	"testing"

	"github.com/powerman/check"
)

func TestPublishSubscribe(tt *testing.T) {
	t := check.T(tt)
	s := NewServer()
	topic := "t1"
	sub := "s1"
	msg := []byte(`{"foo":"bar"}`)
	s.Subscribe(topic, sub)
	s.Publish(topic, msg)
	polledMsg, err := s.Poll(topic, sub)
	t.Nil(err)
	if t.NotNil(polledMsg) {
		bytes.Equal(polledMsg, msg)
	}
	polledMsg, err = s.Poll(topic, sub)
	t.Nil(err)
	t.Nil(polledMsg)
}

func TestPublishSubscribeReceiveMsgsOnlyAfterSubscription(tt *testing.T) {
	t := check.T(tt)
	s := NewServer()
	topic := "t1"
	sub := "s1"
	msg1 := []byte(`{"foo":"bar"}`)
	msg2 := []byte(`{"foo":"baz"}`)
	msg3 := []byte(`{"foo":"qux"}`)
	s.Publish(topic, msg1)
	s.Subscribe(topic, sub)
	s.Publish(topic, msg2)
	s.Publish(topic, msg3)
	polledMsg, err := s.Poll(topic, sub)
	t.Nil(err)
	if t.NotNil(polledMsg) {
		bytes.Equal(polledMsg, msg2)
	}
	polledMsg, err = s.Poll(topic, sub)
	t.Nil(err)
	if t.NotNil(polledMsg) {
		bytes.Equal(polledMsg, msg3)
	}
	polledMsg, err = s.Poll(topic, sub)
	t.Nil(err)
	t.Nil(polledMsg)
}

func TestDoubleSubscribeDoesntLostMessages(tt *testing.T) {
	t := check.T(tt)
	s := NewServer()
	topic := "t1"
	sub := "s1"
	msg1 := []byte(`{"foo":"bar"}`)
	msg2 := []byte(`{"foo":"baz"}`)
	s.Subscribe(topic, sub)
	s.Publish(topic, msg1)
	s.Subscribe(topic, sub)
	s.Publish(topic, msg2)
	polledMsg, err := s.Poll(topic, sub)
	t.Nil(err)
	if t.NotNil(polledMsg) {
		bytes.Equal(polledMsg, msg1)
	}
	polledMsg, err = s.Poll(topic, sub)
	t.Nil(err)
	if t.NotNil(polledMsg) {
		bytes.Equal(polledMsg, msg2)
	}
	polledMsg, err = s.Poll(topic, sub)
	t.Nil(err)
	t.Nil(polledMsg)
}

func TestPollNoSubscribe(tt *testing.T) {
	t := check.T(tt)
	s := NewServer()
	topic := "t1"
	sub := "s1"
	msg, err := s.Poll(topic, sub)
	t.Equal(err, ErrSubscriptionNotFound)
	t.Nil(msg)
}

func TestPollNoSubscribeButClientExist(tt *testing.T) {
	t := check.T(tt)
	s := NewServer()
	topic := "t1"
	topic2 := "t2"
	sub := "s1"
	s.Subscribe(topic2, sub)
	msg, err := s.Poll(topic, sub)
	t.Equal(err, ErrSubscriptionNotFound)
	t.Nil(msg)
}

func TestUnsubscribe(tt *testing.T) {
	t := check.T(tt)
	s := NewServer()
	topic1 := "t1"
	topic2 := "t2"
	sub1 := "s1"
	sub2 := "s2"
	s.Subscribe(topic1, sub1)
	s.Subscribe(topic2, sub1)
	s.Subscribe(topic1, sub2)
	s.Subscribe(topic2, sub2)

	s.Unsubscribe(topic2, sub1)
	s.Unsubscribe(topic1, sub2)

	msg1 := []byte(`{"foo":"bar"}`)
	msg2 := []byte(`{"foo":"baz"}`)

	s.Publish(topic1, msg1)
	s.Publish(topic2, msg2)

	msg, err := s.Poll(topic2, sub1)
	t.Equal(err, ErrSubscriptionNotFound)
	t.Nil(msg)
	msg, err = s.Poll(topic1, sub1)
	t.Nil(err)
	if t.NotNil(msg) {
		bytes.Equal(msg, msg1)
	}

	msg, err = s.Poll(topic1, sub2)
	t.Equal(err, ErrSubscriptionNotFound)
	t.Nil(msg)
	msg, err = s.Poll(topic2, sub2)
	t.Nil(err)
	if t.NotNil(msg) {
		bytes.Equal(msg, msg1)
	}

	s.Unsubscribe(topic1, sub1)
	s.Unsubscribe(topic2, sub2)

	_, exist := s.clients[subscriberName(sub1)]
	t.False(exist)
	_, exist = s.clients[subscriberName(sub2)]
	t.False(exist)

	_, exist = s.topics[topicName(topic1)]
	t.False(exist)
	_, exist = s.topics[topicName(topic2)]
	t.False(exist)

	msg, err = s.Poll(topic1, sub1)
	t.Equal(err, ErrSubscriptionNotFound)
	t.Nil(msg)

	msg, err = s.Poll(topic2, sub2)
	t.Equal(err, ErrSubscriptionNotFound)
	t.Nil(msg)
}

// In this test we start 100 subscribers.
// Each listens on 10 topics.
// After we start 100 publishers.
// Each publish 50 messages on 50 topics.
// Only 10 topics exists. Every publisher create only 1 message per existing topic.
// So every subscriber on topic should receive only 100 messages
// and total 1000 messages(100 messages * 10 topics).
func TestManySubscribesWithManyPublishers(tt *testing.T) {
	t := check.T(tt)
	s := NewServer()
	var swg sync.WaitGroup
	var subs []string
	var topics []string
	for i := 0; i < 10; i++ {
		topics = append(topics, fmt.Sprintf("topic%d", i))
	}
	for i := 0; i < 100; i++ {
		sub := fmt.Sprintf("sub%d", i)
		subs = append(subs, sub)
		for j := 0; j < 10; j++ {
			swg.Add(1)
			j := j
			go func() {
				s.Subscribe(topics[j], sub)
				swg.Done()
			}()
		}
	}
	swg.Wait()

	newMsg := func(i int) []byte {
		return []byte(fmt.Sprintf(`{"foo":"bar%d"}`, i))
	}

	var pwg sync.WaitGroup
	signal := make(chan struct{})
	for i := 0; i < 100; i++ {
		for j := 0; j < 50; j++ {
			topic := fmt.Sprintf("topic%d", j)
			pwg.Add(1)
			msg := newMsg(j)
			go func() {
				<-signal
				s.Publish(topic, msg)
				pwg.Done()
			}()
		}
	}
	close(signal)
	pwg.Wait()
	for _, sub := range subs {
		var subMsgCounter int
		for i, topic := range topics {
			var topicMsgCoutner int
			func() {
				for {

					msg, err := s.Poll(topic, sub)
					t.Nil(err)
					if msg == nil {
						return
					}
					topicMsgCoutner++
					expected := newMsg(i)
					t.True(bytes.Equal(msg, expected), "%s %s %s", sub, topic, string(msg))
				}
			}()
			t.Equal(topicMsgCoutner, 100)
			subMsgCounter += topicMsgCoutner
		}
		t.Equal(subMsgCounter, 1000)
	}
}
