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
func (g *Generator) Start(outerWG *sync.WaitGroup) {
	g.ctx, g.cancel = context.WithTimeout(g.parentctx, g.timeout)
	ticker := time.NewTicker(g.sendPeriod)
	go func() {
		g.logger.Println("generation started")

		defer outerWG.Done()
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
		for {
			select {
			case <-g.ctx.Done():
				g.logger.Println("DONE")
				wg.Wait()
				return
			case <-ticker.C:
				g.logger.Println("ticker event, generating new data")
				stoptaskFn()
				stoptaskFn = g.newTask(g.ctx, &wg, g.out)
			}
		}
	}()
}

func (g *Generator) Outbound() <-chan Data {
	return g.out
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
				ds.InitValue += rand.Intn(ds.MaxChangeStep)
				ds.count++
			}
		}
	}()
	return cancel
}
