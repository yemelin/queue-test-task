package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
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