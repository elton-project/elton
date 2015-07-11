package api

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	Master   MasterConfig   `toml:"master"`
	Slave    SlaveConfig    `toml:"slave"`
	Masters  []MasterConfig `toml:"masters"`
	Database DBConfig       `toml:"database"`
}

type MasterConfig struct {
	Name     string `toml:"name"`
	HostName string `toml:"hostname"`
}

type SlaveConfig struct {
	GrpcHostName   string `toml:"grpc_hostname"`
	HttpHostName   string `toml:"http_hostname"`
	MasterHostName string `toml:"master_hostname"`
	Dir            string `toml:"dir"`
}

type DBConfig struct {
	DBPath string `toml:"dbpath"`
}

func Load(path string) (Config, error) {
	var conf Config
	_, err := toml.DecodeFile(path, &conf)
	if err != nil {
		return Config{}, err
	}

	return conf, nil
}
