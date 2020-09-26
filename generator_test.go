// +build async

package main

import (
	"testing"
	"time"
)

func createGen() *Generator {
	return &Generator{
		timeout:    1000 * time.Millisecond,
		sendPeriod: 100 * time.Millisecond,
		out:        make(chan Data),
		dataSources: []DataSource{
			{ID: "data_0", InitValue: 10, MaxChangeStep: 100},
			{ID: "data_1", InitValue: 10, MaxChangeStep: 10},
			{ID: "data_2", InitValue: 10, MaxChangeStep: 1},
		},
	}
}

func TestSimple(t *testing.T) {

	t.Run("should be stopped by force", func(t *testing.T) {
		g := createGen()
		var stoppedByForce bool
		go func() {
			time.Sleep(200 * time.Millisecond)
			g.Stop()
			stoppedByForce = true
		}()
		_, done := g.Start()
		<-done
		if !stoppedByForce {
			t.Fatalf("not stopped by force")
		}
	})

	t.Run("should not block sending (part of the values is read)", func(t *testing.T) {
		g := createGen()
		var receiver []Data
		var stoppedByForce bool
		go func() {
			time.Sleep(2000 * time.Millisecond)
			g.Stop()
			stoppedByForce = true
		}()
		in, done := g.Start()
		for i := 0; i < 5; i++ {
			receiver = append(receiver, <-in)
		}
		<-done
		if stoppedByForce {
			t.Fatalf("should stop by itself")
		}
	})

	t.Run("should not block sending (receiver sleeps)", func(t *testing.T) {
		g := createGen()
		var receiver []Data
		var stoppedByForce bool
		go func() {
			time.Sleep(2000 * time.Millisecond)
			g.Stop()
			stoppedByForce = true
		}()
		in, done := g.Start()
		for v := range in {
			receiver = append(receiver, v)
			time.Sleep(200 * time.Millisecond)
		}
		<-done
		if stoppedByForce {
			t.Fatalf("should stop by itself")
		}
	})

	t.Run("should read all values)", func(t *testing.T) {
		g := createGen()
		var receiver []Data
		var stoppedByForce bool
		go func() {
			time.Sleep(2000 * time.Millisecond)
			g.Stop()
			stoppedByForce = true
		}()
		in, done := g.Start()
		for v := range in {
			receiver = append(receiver, v)
		}
		<-done
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
