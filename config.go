package main

import (
	"encoding/json"
	"errors"
	"fmt"
)

type AppConfig struct {
	Generators []struct {
		TimeoutS    int `json:"timeout_s"`
		SendPeriodS int `json:"send_period_s"`
		DataSources []struct {
			ID            string `json:"id"`
			InitValue     int    `json:"init_value"`
			MaxChangeStep int    `json:"max_change_step"`
		} `json:"data_sources"`
	} `json:"generators"`
	Agregators []struct {
		SubIds          []string `json:"sub_ids"`
		AgregatePeriodS int      `json:"agregate_period_s"`
	} `json:"agregators"`
	Queue struct {
		Size int `json:"size"`
	} `json:"queue"`
	StorageType int `json:"storage_type"`
}

func LoadConfig(b[]byte) (*AppConfig, error) {
	c := AppConfig{}
	err := json.Unmarshal(b, &c)
	if err != nil {
		return nil, err
	}
	if err := Validate(&c); err != nil {
		return nil, err
	}
	return &c, nil
}

// TODO: use validation struct tags
func Validate(c *AppConfig) error {
	if c.StorageType < 0 || c.StorageType >1 {
		return errors.New(fmt.Sprintf("invalid storage type %d", c.StorageType))
	}
	return nil
}