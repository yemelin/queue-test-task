package main

import (
	"sync"
	"time"
)

type Subscription struct {
	id string
	in <-chan Data
}

type Aggregator struct {
	topics      []string
	aggregators []aggregator
	period      time.Duration
	storage     *Storage
	logger      *Logger
}

func NewAggregator(topics []string, period int, storage *Storage) *Aggregator {
	a := &Aggregator{
		topics:  topics,
		period:  time.Duration(period) * time.Second,
		storage: storage,
		logger:  NewLogger("aggregator"),
	}
	return a
}

func (a *Aggregator) Topics() []string {
	return a.topics
}

func (a *Aggregator) Subscribe(subscription Subscription) {
	a.aggregators = append(a.aggregators, aggregator{
		id:     subscription.id,
		in:     subscription.in,
		dump:   make(chan struct{}),
		s:      a.storage,
		logger: NewLogger("agg_" + subscription.id),
	})
}

func (a *Aggregator) Start(outerWG *sync.WaitGroup) {
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
	go func() {
		defer func() {
			ticker.Stop()
			outerWG.Done()
		}()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				for _, child := range a.aggregators {
					<-child.dump
				}
			}
		}
	}()
	go func() {
		a.logger.Println("children monitor started")
		wg.Wait()
		done <- struct{}{}
		a.logger.Println("children monitor exited")
	}()
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
			a.logger.Debugf("Sent to storage: Avg. %s (%d values): %f", a.id, len(a.avg.vals), a.avg.Value())
			_ = a.s.Store(Record{a.id, len(a.avg.vals), a.avg.Value()})
			a.count++
			a.avg.Reset()
			a.avg.Update(d.Value)
		default:
			a.avg.Update(d.Value)
		}
	}
	close(a.dump)
	if len(a.avg.vals) > 0 {
		a.logger.Debugf("Sent to storage: Avg. %s (%d values): %f", a.id, len(a.avg.vals), a.avg.Value())
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
