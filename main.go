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
	conf := config.Generators[0]
	dataSources := make([]DataSource, len(conf.DataSources))
	for i, ds := range conf.DataSources {
		dataSources[i] = DataSource{
			ID:            ds.ID,
			InitValue:     ds.InitValue,
			MaxChangeStep: ds.MaxChangeStep,
		}
	}
	ctx, shutDown := context.WithCancel(context.Background())
	g := &Generator{
		parentctx:   ctx,
		timeout:     100 * time.Duration(conf.TimeoutS) * time.Millisecond,
		sendPeriod:  100 * time.Duration(conf.SendPeriodS) * time.Millisecond,
		out:         make(chan Data),
		dataSources: dataSources,
		logger:      NewLogger("Generator"),
	}

	q := NewQueue(10)
	q.AddPublisher(g.out)
	var subscriptions []Subscription
	for _, topic := range []string{"data_1", "data_2", "data_3"} {
		subscriptions = append(subscriptions, Subscription{topic, q.Subscription(topic)})
	}
	w := os.Stdout
	if config.StorageType == 1 {
		w, err = os.Open("data.txt")
		if err != nil {
			logger.Fatalf("failed to create storage file %s: %v", "data.txt", err)
		}
	}
	storage := &Storage{w: w, logger: NewLogger("Storage")}
	_ = NewAggregator(subscriptions, config.Agregators[0].AgregatePeriodS, storage)

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

	_, done := g.Start()

	<-done
	logger.Println("received DONE from generator, closing queue")
	q.Close()
	err = storage.Close(5)
}

func InitApp(args []string) (*AppConfig, error) {
	f, err := os.Open(os.Args[2])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("eror opening %s; %v", os.Args[2], err))
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
