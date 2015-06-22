package elton

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	Elton    EltonConfig   `toml:"elton"`
	Masters  []EltonConfig `toml:"master"`
	Database DBConfig      `toml:"database"`
}

type EltonConfig struct {
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
