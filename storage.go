package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)

// TODO: generate record ID
type Record struct {
	dataID    string
	numValues int
	avg       float64
}

func (r Record) String() string {
	return fmt.Sprintf("data type %s, average of %d values: %f", r.dataID, r.numValues, r.avg)
}

type Storage struct {
	w  io.WriteCloser
	wg sync.WaitGroup
}

func (s *Storage) Store(r Record) Result {
	var wg sync.WaitGroup
	var err error
	wg.Add(1)
	s.wg.Add(1)
	fmt.Println("STORE ", r)
	go func() {
		_, err = io.WriteString(s.w, fmt.Sprintf("%v -- %v\n", time.Now(), r))
		wg.Done()
		s.wg.Done()
	}()
	return Result{err, &wg}
}

func (s *Storage) Close(d int) error {
	fmt.Println("closing storage")
	done := make(chan struct{})
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(d)*time.Second)
	defer cancel()
	go func() {
		s.wg.Wait()
		// s.w.Close()
		done <- struct{}{}
	}()
	fmt.Println("before select")
	for {
		fmt.Println("inside select")
		select {
		case <-done:
			fmt.Println("storage closed gracefully")
			return nil
		case <-ctx.Done():
			fmt.Println("failed to close Storage gracefully")
			return errors.New("failed to close Storage gracefully")
		}
	}
}

type Result struct {
	err error
	wg  *sync.WaitGroup
}

func (r *Result) Get() error {
	r.wg.Wait()
	return r.err
}
