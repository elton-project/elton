package api

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	Elton    MasterConfig   `toml:"elton"`
	Masters  []MasterConfig `toml:"master"`
	Database DBConfig       `toml:"database"`
}

type MasterConfig struct {
	Name     string `toml:"name"`
	HostName string `toml:"hostname"`
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
