package main

import "fmt"

type Queue struct {
	q           chan Data
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
func (q *Queue) Subscribe(topic string) <-chan Data {
	c, ok := q.subscribers[topic]
	if !ok {
		c = make(chan Data)
		q.subscribers[topic] = c
	}
	return c
}

func (q *Queue) start() {
	go func() {
		for d := range q.q {
			q.subscribers[d.ID] <- d
		}
		fmt.Println("afer the subscriber loop")
	}()
}
