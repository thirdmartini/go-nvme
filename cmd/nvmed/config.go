package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type TargetConfig struct {
	Name            string
	Type            string
	ModelName       string
	SerialNumber    string
	FirmwareVersion string
	UUID            string
	Options         map[string]string
}

type Config struct {
	Targets []*TargetConfig
}

func LoadConfig(name string) (*Config, error) {
	data, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}

	conf := &Config{}
	err = yaml.Unmarshal(data, conf)
	return conf, err
}
