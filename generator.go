package main

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

type DataSource struct {
	ID            string `json:"id"`
	InitValue     int    `json:"init_value"`
	MaxChangeStep int    `json:"max_change_step"`
	count         int
}

type Generator struct {
	parentctx   context.Context
	ctx         context.Context
	cancel      func()
	timeout     time.Duration
	sendPeriod  time.Duration
	out         chan Data
	dataSources []DataSource
	logger      *Logger
}

// TODO: either make resettable, or prohibit re-use
func (g *Generator) Start() (out <-chan Data, d <-chan struct{}) {
	done := make(chan struct{})
	g.ctx, g.cancel = context.WithTimeout(g.parentctx, g.timeout)
	ticker := time.NewTicker(g.sendPeriod)
	go func() {
		g.logger.Println("generation started")
		defer func() { done <- struct{}{} }()
		defer func() {
			for _, ds := range g.dataSources {
				g.logger.Printf("sent %d values of type %s\n", ds.count, ds.ID)
			}
		}()
		defer close(g.out)
		defer g.cancel()
		defer ticker.Stop()

		var wg sync.WaitGroup
		stoptaskFn := g.newTask(g.ctx, &wg, g.out)
		// fmt.Println("before ticker loop")
		for {
			select {
			case <-g.ctx.Done():
				g.logger.Println("DONE")
				wg.Wait()
				return
			case <-ticker.C:
				// fmt.Println("ticker event")
				stoptaskFn()
				stoptaskFn = g.newTask(g.ctx, &wg, g.out)
			}
		}
	}()
	return g.out, done
}

func (g *Generator) Stop() {
	if g.cancel != nil {
		g.logger.Println("stop signal received")
		g.cancel()
	}
}

func (g *Generator) newTask(ctx context.Context, wg *sync.WaitGroup, ch chan Data) (cancelFn func()) {
	childCtx, cancel := context.WithCancel(ctx)
	// update all datasources. if updated value is not sent it's bypassed, i.e. discarded
	for i := range g.dataSources {
		ds := &g.dataSources[i]
		ds.InitValue += rand.Intn(ds.MaxChangeStep)
	}
	dsNums := rand.Perm(len(g.dataSources)) //shuffle
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()
		for _, i := range dsNums {
			ds := &g.dataSources[i]
			select {
			case <-childCtx.Done():
				return
			case g.out <- Data{ds.ID, ds.InitValue}:
				// fmt.Println("value sent")
				ds.InitValue += rand.Intn(ds.MaxChangeStep)
				ds.count++
			}
		}
	}()
	return cancel
}
