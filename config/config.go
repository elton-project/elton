package config

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	Server ServerConfig
	Db     DbConfig
}

type ServerConfig struct {
	Port  string        `toml:"port"`
	Slave []SlaveServer `toml:"slave"`
}

type DbConfig struct {
	Host string `toml:"host"`
	Port string `toml:"port"`
	User string `toml:"user"`
	Pass string `toml:"pass"`
}

type SlaveServer struct {
	Weight int    `toml:"weight"`
	Host   string `toml:"host"`
	Port   string `toml:"port"`
}

func Load(path string) (Config, error) {
	var conf Config
	_, err := toml.DecodeFile(path, &conf)
	if err != nil {
		return Config{}, err
	}
	return conf, nil
}
