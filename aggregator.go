package main

import (
	"sync"
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
	// wg          sync.WaitGroup
	done   chan struct{}
	logger *Logger
}

func NewAggregator(subscriptions []Subscription, period int, storage *Storage) (*Aggregator, chan struct{}) {
	aggregators := make([]aggregator, len(subscriptions))
	for i, subscription := range subscriptions {
		aggregators[i].id = subscription.id
		aggregators[i].in = subscription.in
		aggregators[i].dump = make(chan struct{})
		aggregators[i].s = storage
		aggregators[i].logger = NewLogger("Agg_" + subscription.id)
	}
	done := make(chan struct{})
	a := &Aggregator{
		aggregators: aggregators,
		period:      100 * time.Duration(period) * time.Millisecond,
		done:        done,
		logger:      NewLogger("Aggregator"),
	}
	a.run()
	return a, done
}

func (a *Aggregator) run() {
	var wg sync.WaitGroup
	wg.Add(len(a.aggregators))
	for _, child := range a.aggregators {
		child := child
		go func() {
			child.run()
			wg.Done()
			a.logger.Println("child exited")
		}()
	}
	done := make(chan struct{})
	ticker := time.NewTicker(a.period)
	go func(done chan struct{}) {
		defer func() {
			ticker.Stop()
			a.done <- struct{}{}
		}()
		for {
			select {
			case <-done:
				a.logger.Printf("done received")
				return
			case <-ticker.C:
				a.logger.Println("aggregator ticker event")
				for _, child := range a.aggregators {
					<-child.dump
				}
			}
		}
	}(done)
	go func(done chan struct{}) {
		a.logger.Println("children monitor started")
		wg.Wait()
		a.logger.Println("children monitor exited")
		done <- struct{}{}
		a.logger.Println("children monitor really exited")
	}(done)
}

type aggregator struct {
	id     string
	in     <-chan Data
	dump   chan struct{}
	avg    average
	s      *Storage
	count  int
	logger *Logger
}

func (a *aggregator) run() {
	a.logger.Printf("child aggregator for %s started\n", a.id)
	for d := range a.in {
		select {
		case a.dump <- struct{}{}:
			a.logger.Printf("Sent to storage: Avg. %s (%d values): %f", a.id, len(a.avg.vals), a.avg.Value())
			_ = a.s.Store(Record{a.id, len(a.avg.vals), a.avg.Value()})
			a.count++
			a.avg.Reset()
			a.avg.Update(d.Value)
		default:
			a.avg.Update(d.Value)
		}
	}
	close(a.dump)
	a.logger.Printf("child loop done")
	if len(a.avg.vals) > 0 {
		// a.logger.Printf("Sent to storage: Avg. %s (%d values): %f", a.id, len(a.avg.vals), a.avg.Value())
		_ = a.s.Store(Record{a.id, len(a.avg.vals), a.avg.Value()})
		a.count++
	}
	a.logger.Printf("total: %d aggregations for %s, exiting", a.count, a.id)
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
