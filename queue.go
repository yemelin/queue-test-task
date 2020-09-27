package main

import (
	"fmt"
	"sync"
)

type Queue struct {
	q           chan Data
	lock        sync.RWMutex
	subscribers map[string]chan Data
}

func NewQueue(cap int) *Queue {
	queue := &Queue{q: make(chan Data, cap), subscribers: make(map[string]chan Data)}
	queue.start()
	return queue
}

func (q *Queue) AddPublisher(p <-chan Data) {
	go func() {
		fmt.Println("add publisher")
		for d := range p {
			fmt.Printf("trying to consume %s: %d\n", d.ID, d.Value)
			select {
			case q.q <- d:
				fmt.Println("value written to queue")
			}
		}
		fmt.Println("afer the loop")
	}()
}

// TODO: handle subscription attempts to the same topic
func (q *Queue) Subscription(topic string) <-chan Data {
	q.lock.RLock()
	c, ok := q.subscribers[topic]
	q.lock.RUnlock()
	if !ok {
		q.lock.Lock()
		c, ok = q.subscribers[topic]
		if !ok {
			c = make(chan Data)
			q.subscribers[topic] = c
		}
		q.lock.Unlock()
		fmt.Printf("subscription to %s created\n", topic)
	}
	return c
}

func (q *Queue) start() {
	go func() {
		fmt.Println("queue started")
		for d := range q.q {
			q.lock.RLock()
			s, ok := q.subscribers[d.ID]
			q.lock.RUnlock()
			if ok {
				s <- d
				fmt.Printf("%v read from queue\n", d)
			} else {
				fmt.Printf("%v discarded\n", d)
			}
		}
		fmt.Println("afer the subscriber loop")
	}()
}
