package config

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	Proxy  ProxyConfig
	Server []ServerConfig
	DB     DBConfig
}

type ProxyConfig struct {
	Port string `toml:"port"`
}

type ServerConfig struct {
	Weight int    `toml:"weight"`
	Host   string `toml:"host"`
	Port   string `toml:"port"`
}

type DBConfig struct {
	User   string `toml:"user"`
	Pass   string `toml:"pass"`
	Host   string `toml:"host"`
	Port   string `toml:"port"`
	DBName string `toml:"dbname"`
}

func Load(path string) (Config, error) {
	var conf Config
	_, err := toml.DecodeFile(path, &conf)
	if err != nil {
		return Config{}, err
	}
	return conf, nil
}
