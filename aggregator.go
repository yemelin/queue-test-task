package main

import (
	"fmt"
	"time"
)

type Subscription struct {
	id string
	in <-chan Data
}

// TODO: add stopping mechanism - context
type Aggregator struct {
	aggregators []aggregator
	period      time.Duration
}

func NewAggregator(subscriptions []Subscription, period int) *Aggregator {
	aggregators := make([]aggregator, len(subscriptions))
	for i, subscription := range subscriptions {
		aggregators[i].id = subscription.id
		aggregators[i].in = subscription.in
		aggregators[i].dump = make(chan struct{})
	}
	fmt.Println(aggregators)
	a := &Aggregator{aggregators: aggregators, period: time.Duration(period) * time.Second}
	a.run()
	return a
}

func (a *Aggregator) run() {	
	fmt.Printf("starting %d children\n", len(a.aggregators))
	for _, child := range a.aggregators {
		child := child
		go child.run()
	}
	ticker := time.NewTicker(a.period)
	go func() {
		select {
		case <-ticker.C:
			for _, child := range a.aggregators {
				child.dump <- struct{}{}
			}
		}
	}()
}

type aggregator struct {
	id   string
	in   <-chan Data
	dump chan struct{}
	avg  average
}

func (a *aggregator) run() {
	fmt.Printf("child %s started\n", a.id)
	for d := range a.in {
		select {
		case <-a.dump:
			fmt.Printf("Avg. %s (%d values): %f", a.id, len(a.avg.vals), a.avg.Value())
			a.avg.Reset()
			a.avg.Update(d.Value)
		default:
			a.avg.Update(d.Value)
		}
	}
}

type average struct {
	vals []int
}

func (a *average) Value() float64 {
	var sum int
	for _, v := range a.vals {
		sum += v
	}
	return float64(sum) / float64(len(a.vals))
}

func (a *average) Reset() {
	a.vals = make([]int, 0)
}

func (a *average) Update(i int) {
	a.vals = append(a.vals, i)
}
