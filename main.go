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

const storageTimeout = 5

func main() {
	if len(os.Args) != 3 || os.Args[1] != "--config" {
		fmt.Printf("usage: %s --config <filename>", os.Args[0])
		os.Exit(0)
	}
	logger := NewLogger("main")
	config, err := InitApp(os.Args)
	if err != nil {
		logger.Fatalln(err)
	}

	q := NewQueue(config.Queue.Size)
	generators, shutDown := createGenerators(config)
	storage, err := createStorage(config)
	if err != nil {
		logger.Fatalf("failed to create storage file %s: %v", "data.txt", err)
	}
	aggregators := createAggregators(config, storage)
	pubsub := &PubSubManager{q: q}

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

	pubsub.AddSubscribers(aggregators)
	pubsub.AddPublishers(generators)
	pubsub.Wait()
	err = storage.Close(storageTimeout)
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

func createGenerators(config *AppConfig) (generators []Publisher, cancel func()) {
	ctx, shutDown := context.WithCancel(context.Background())
	generators = make([]Publisher, len(config.Generators))
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
		generators[i] = &Generator{
			parentctx:   ctx,
			timeout:     time.Duration(conf.TimeoutS) * time.Second,
			sendPeriod:  time.Duration(conf.SendPeriodS) * time.Second,
			out:         make(chan Data),
			dataSources: dataSources,
			logger:      NewLogger("generator"),
		}
	}
	return generators, shutDown
}

func createStorage(config *AppConfig) (*Storage, error) {
	w := os.Stdout
	if config.StorageType == 1 {
		fname := "data.txt"
		srcDir := os.Getenv("SRC_DIR")
		if srcDir != "" {
			fname = fmt.Sprintf("%s/%s", srcDir, fname)
		}
		var err error
		w, err = os.Create(fname)
		if err != nil {
			return nil, err
		}
	}
	return &Storage{w: w, logger: NewLogger("Storage")}, nil
}

func createAggregators(config *AppConfig, storage *Storage) []Subscriber {
	ret := make([]Subscriber, len(config.Agregators))
	for i, c := range config.Agregators {
		ret[i] = NewAggregator(c.SubIds, c.AgregatePeriodS, storage)
	}
	return ret
}
