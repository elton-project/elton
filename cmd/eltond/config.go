package main

import (
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
)

type EnvConfig struct {
	Roles []string `required:"true"`
	//ControllerListenAddr tcpAddr `split_words:"true"`
	//Controllers          tcpAddrs
}

func loadFromEnvironment() (*EnvConfig, error) {
	conf := &EnvConfig{}
	err := envconfig.Process("ELTON", conf)
	if err != nil {
		return nil, err
	}

	zap.S().With(
		"roles", conf.Roles,
	).Info("loaded configuration from environment")
	return conf, nil
}
