package main

import (
	"fmt"
	"sync"
)

type Publisher interface {
	Outbound() <-chan Data
	Start(*sync.WaitGroup)
}

type Subscriber interface {
	Topics() []string
	Subscribe(Subscription)
	Start(*sync.WaitGroup)
}

type PubSubManager struct {
	q     *Queue
	pubwg sync.WaitGroup
	subwg sync.WaitGroup
}

func (m *PubSubManager) AddPublishers(publishers []Publisher) {
	for _, publisher := range publishers {
		m.q.AddPublisher(publisher.Outbound())
	}
	for _, publisher := range publishers {
		m.pubwg.Add(1)
		publisher.Start(&m.pubwg)
	}
}

// add before publishers
func (m *PubSubManager) AddSubscribers(subscribers []Subscriber) {
	for _, subscriber := range subscribers {
		for _, topic := range subscriber.Topics() {
			subscription := m.q.Subscription(topic)
			subscriber.Subscribe(Subscription{topic, subscription})
		}
	}
	for _, subscriber := range subscribers {
		m.subwg.Add(1)
		subscriber.Start(&m.subwg)
	}
}

func (m *PubSubManager) Wait() {
	fmt.Println("waiting for queue")
	m.q.Wait()
	fmt.Println("waiting for publishers")
	m.pubwg.Wait()
	fmt.Println("waiting for subscribers")
	m.subwg.Wait()
}
