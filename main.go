package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if len(os.Args) != 3 || os.Args[1] != "--config" {
		fmt.Printf("usage: %s --config <filename>", os.Args[0])
		os.Exit(0)
	}
	logger := NewLogger("Main")
	config, err := InitApp(os.Args)
	if err != nil {
		logger.Fatalln(err)
	}
	ctx, shutDown := context.WithCancel(context.Background())
	q := NewQueue(config.Queue.Size)
	generators := make([]Generator, len(config.Generators))
	for i := 0; i < len(config.Generators); i++ {
		conf := config.Generators[i]
		dataSources := make([]DataSource, len(conf.DataSources))
		for i, ds := range conf.DataSources {
			dataSources[i] = DataSource{
				ID:            ds.ID,
				InitValue:     ds.InitValue,
				MaxChangeStep: ds.MaxChangeStep,
			}
		}
		generators[i] = Generator{
			parentctx:   ctx,
			timeout:     100 * time.Duration(conf.TimeoutS) * time.Millisecond,
			sendPeriod:  100 * time.Duration(conf.SendPeriodS) * time.Millisecond,
			out:         make(chan Data),
			dataSources: dataSources,
			logger:      NewLogger("Generator"),
		}

		q.AddPublisher(generators[i].out)
	}

	w := os.Stdout
	if config.StorageType == 1 {
		w, err = os.Create("data.txt")
		if err != nil {
			logger.Fatalf("failed to create storage file %s: %v", "data.txt", err)
		}
	}
	storage := &Storage{w: w, logger: NewLogger("Storage")}

	aggDone := make([]<-chan struct{}, len(config.Agregators))
	for i := range config.Agregators {
		var subscriptions []Subscription
		for _, topic := range config.Agregators[i].SubIds {
			subscriptions = append(subscriptions, Subscription{topic, q.Subscription(topic)})
		}
		_, aggDone[i] = NewAggregator(subscriptions, config.Agregators[0].AgregatePeriodS, storage)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(
		stop,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	go func() {
		<-stop
		logger.Println("caught stop signal")
		shutDown()
	}()
	done := make([]<-chan struct{}, len(generators))
	for i := range generators {
		_, done[i] = generators[i].Start()
	}
	for i := 0; i < len(done); i++ {
		<-done[i]
	}

	logger.Println("received DONE from generator, waiting for aggregators")
	for i := 0; i < len(aggDone); i++ {
		<-aggDone[i]
	}
	err = storage.Close(5)
}

func InitApp(args []string) (*AppConfig, error) {
	f, err := os.Open(os.Args[2])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error opening %s; %v", os.Args[2], err))
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error reading from %s: %v", os.Args[2], err))
	}
	config, err := LoadConfig(b)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error loading config; %v", err))
	}
	return config, err
}
