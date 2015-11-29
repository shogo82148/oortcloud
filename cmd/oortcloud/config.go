package main

import (
	"errors"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

// Config is the definition of the configuration YAML file
type Config struct {
	// API is the settings for the api server
	API *struct {
		Host                string   `yaml:"host"`
		Port                string   `yaml:"port"`
		Sock                string   `yaml:"sock"`
		Callback            []string `yaml:"callback"`
		MaxIdleConnsPerHost int      `yaml:"max_idle_conns_per_host"`
	} `yaml:"api"`

	// Websocket is the settings for the websocket server
	Websocket *struct {
		Host   string `yaml:"host"`
		Port   string `yaml:"port"`
		Sock   string `yaml:"sock"`
		Binary bool   `yaml:"binary"`
	} `yaml:"websocket"`
}

func LoadConfig(filename string) (*Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = yaml.Unmarshal(b, config)
	if err != nil {
		return nil, err
	}

	if config.API == nil {
		return nil, errors.New("missing API configuration")
	}

	return config, nil
}
