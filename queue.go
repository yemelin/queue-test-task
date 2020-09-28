package main

import (
	"sync"
)

type Queue struct {
	q               chan Data
	lock            sync.RWMutex
	subscribers     map[string]chan Data
	wg              sync.WaitGroup
	noNewPublishers chan struct{}
	logger          *Logger
}

func NewQueue(cap int) *Queue {
	queue := &Queue{
		q:               make(chan Data, cap),
		subscribers:     make(map[string]chan Data),
		noNewPublishers: make(chan struct{}),
		logger:          NewLogger("Queue"),
	}
	queue.start()
	return queue
}

func (q *Queue) AddPublisher(p <-chan Data) {
	q.wg.Add(1)
	go func() {
		q.logger.Println("publisher added")
		for d := range p {
			q.logger.Debugf("trying to consume %s: %d\n", d.ID, d.Value)
			select {
			case q.q <- d:
				q.logger.Debugf("value written to queue")
			}
		}
		q.wg.Done()
		q.logger.Println("publishing loop exited")
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
		q.logger.Printf("subscription to %s created\n", topic)
	}
	return c
}

func (q *Queue) start() {
	go func() {
		q.logger.Println("queue started")
		for d := range q.q {
			q.lock.RLock()
			s, ok := q.subscribers[d.ID]
			q.lock.RUnlock()
			if ok {
				s <- d
				q.logger.Debugf("%v read from queue\n", d)
			} else {
				q.logger.Debugf("%v discarded\n", d)
			}
		}
		q.logger.Println("queue exhausted")
		q.lock.RLock()
		for k, c := range q.subscribers {
			q.logger.Println("closing topic ", k)
			close(c)
		}
		q.lock.RUnlock()
	}()
	go func() {
		q.logger.Println("publishing exhaustion monitor started")
		<-q.noNewPublishers
		q.logger.Println("waiting for exhaustion")
		q.wg.Wait()
		q.logger.Println("all publishers exhausted, closing queue")
		close(q.q)
	}()
}

// accept no more publishers
func (q *Queue) Wait() {
	q.logger.Println("sending no more pubishers signal")
	close(q.noNewPublishers)
	q.logger.Println("no more pubishers signal sent!")
}

func (q *Queue) Close() {
	close(q.q)
}
