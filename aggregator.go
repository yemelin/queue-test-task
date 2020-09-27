package main

import (
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

func NewAggregator(subscriptions []Subscription, period int, storage *Storage) *Aggregator {
	aggregators := make([]aggregator, len(subscriptions))
	for i, subscription := range subscriptions {
		aggregators[i].id = subscription.id
		aggregators[i].in = subscription.in
		aggregators[i].dump = make(chan struct{})
		aggregators[i].s = storage
		aggregators[i].logger = NewLogger("Agg_" + subscription.id)
	}
	// fmt.Println(aggregators)
	a := &Aggregator{aggregators: aggregators, period: 100 * time.Duration(period) * time.Millisecond}
	a.run()
	return a
}

func (a *Aggregator) run() {
	// fmt.Printf("starting %d children\n", len(a.aggregators))
	for _, child := range a.aggregators {
		child := child
		go child.run()
	}
	ticker := time.NewTicker(a.period)
	// TODO: stop this thread!
	go func() {
		for {
			select {
			case <-ticker.C:
				// fmt.Println("aggregator ticker event")
				for _, child := range a.aggregators {
					child.dump <- struct{}{}
				}
			}
		}
	}()
}

type aggregator struct {
	id   string
	in   <-chan Data
	dump chan struct{}
	avg  average
	s    *Storage

	logger *Logger
}

func (a *aggregator) run() {
	a.logger.Printf("child %s started\n", a.id)
	for d := range a.in {
		select {
		case <-a.dump:
			// fmt.Printf("Sent to storage: Avg. %s (%d values): %f", a.id, len(a.avg.vals), a.avg.Value())
			_ = a.s.Store(Record{a.id, len(a.avg.vals), a.avg.Value()})
			a.avg.Reset()
			a.avg.Update(d.Value)
		default:
			a.avg.Update(d.Value)
		}
	}
	a.logger.Println("after child's loop")
	if len(a.avg.vals) > 0 {
		// fmt.Printf("Sent to storage: Avg. %s (%d values): %f", a.id, len(a.avg.vals), a.avg.Value())
		_ = a.s.Store(Record{a.id, len(a.avg.vals), a.avg.Value()})
	}
	a.logger.Println("child exiting")
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
