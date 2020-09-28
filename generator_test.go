// +build async

package main

import (
	"context"
	"sync"
	"testing"
	"time"
)

func createGen(ctx context.Context) *Generator {
	return &Generator{
		parentctx:  ctx,
		timeout:    1000 * time.Millisecond,
		sendPeriod: 100 * time.Millisecond,
		out:        make(chan Data),
		dataSources: []DataSource{
			{ID: "data_0", InitValue: 10, MaxChangeStep: 100},
			{ID: "data_1", InitValue: 10, MaxChangeStep: 10},
			{ID: "data_2", InitValue: 10, MaxChangeStep: 1},
		},
		logger: NewLogger("TestGen"),
	}
}

func TestGenerator(t *testing.T) {
	t.Run("should be stopped by force", func(t *testing.T) {
		ctx := context.Background()
		g := createGen(ctx)
		var stoppedByForce bool
		go func() {
			time.Sleep(200 * time.Millisecond)
			g.Stop()
			stoppedByForce = true
		}()
		var anyWg sync.WaitGroup
		anyWg.Add(1)
		g.Start(&anyWg)
		anyWg.Wait()
		if !stoppedByForce {
			t.Fatalf("not stopped by force")
		}
	})

	t.Run("should not block sending (part of the values is read)", func(t *testing.T) {
		ctx := context.Background()
		g := createGen(ctx)
		var receiver []Data
		var stoppedByForce bool
		go func() {
			time.Sleep(2000 * time.Millisecond)
			g.Stop()
			stoppedByForce = true
		}()
		var anyWg sync.WaitGroup
		in := g.Outbound()
		anyWg.Add(1)
		g.Start(&anyWg)
		for i := 0; i < 5; i++ {
			receiver = append(receiver, <-in)
		}
		anyWg.Wait()
		if stoppedByForce {
			t.Fatalf("should stop by itself")
		}
	})

	t.Run("should not block sending (receiver sleeps)", func(t *testing.T) {
		ctx := context.Background()
		g := createGen(ctx)
		var receiver []Data
		var stoppedByForce bool
		go func() {
			time.Sleep(2000 * time.Millisecond)
			g.Stop()
			stoppedByForce = true
		}()

		var anyWg sync.WaitGroup
		in := g.Outbound()
		anyWg.Add(1)
		g.Start(&anyWg)
		for v := range in {
			receiver = append(receiver, v)
			time.Sleep(200 * time.Millisecond)
		}
		
		anyWg.Wait()
		if stoppedByForce {
			t.Fatalf("should stop by itself")
		}
	})

	t.Run("should read all values)", func(t *testing.T) {
		ctx := context.Background()
		g := createGen(ctx)
		var receiver []Data
		var stoppedByForce bool
		go func() {
			time.Sleep(2000 * time.Millisecond)
			g.Stop()
			stoppedByForce = true
		}()
		var anyWg sync.WaitGroup
		in := g.Outbound()
		anyWg.Add(1)
		g.Start(&anyWg)
		for v := range in {
			receiver = append(receiver, v)
		}
		anyWg.Wait()
		if stoppedByForce {
			t.Fatalf("should stop by itself")
		}
		// < instead of != because from time to time there are 11 100 ms ticks within 1000 ms period,
		// even though timer starts before the ticker
		if len(receiver) < 30 {
			t.Fatalf("expected %d values, got %d", 30, len(receiver))
		}
	})
}
