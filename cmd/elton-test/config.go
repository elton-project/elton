package main

import "github.com/kelseyhightower/envconfig"

type Config struct {
	ControllerListenAddr string `split_words:"true"`
	ControllerListenPort int    `split_words:"true"`
	Controllers          []string
}

func loadFromEnvironment() (*Config, error) {
	conf := &Config{}
	err := envconfig.Process("ELTON", conf)
	if err != nil {
		return nil, err
	}
	return conf, err
}
