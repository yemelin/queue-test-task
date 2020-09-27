package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"
)

func main() {
	if len(os.Args) != 3 || os.Args[1] != "--config" {
		log.Printf("usage: %s --config <filename>", os.Args[0])
		os.Exit(0)
	}
	config, err := InitApp(os.Args)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(*config)
	conf := config.Generators[0]
	dataSources := make([]DataSource, len(conf.DataSources))
	for i, ds := range conf.DataSources {
		dataSources[i] = DataSource{
			ID:            ds.ID,
			InitValue:     ds.InitValue,
			MaxChangeStep: ds.MaxChangeStep,
		}
	}
	g := &Generator{
		timeout:     100 * time.Duration(conf.TimeoutS) * time.Millisecond,
		sendPeriod:  100 * time.Duration(conf.SendPeriodS) * time.Millisecond,
		out:         make(chan Data),
		dataSources: dataSources,
	}
	go func() {
		time.Sleep(5 * time.Second)
		fmt.Println("stopping generator")
		g.Stop()
	}()

	q := NewQueue(10)
	q.AddPublisher(g.out)
	var subscriptions []Subscription
	for _, topic := range []string{"data_1", "data_2", "data_3"} {
		subscriptions = append(subscriptions, Subscription{topic, q.Subscription(topic)})
	}
	_ = NewAggregator(subscriptions, 10)

	_, done := g.Start()
	<-done
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
