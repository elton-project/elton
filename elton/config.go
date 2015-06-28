package elton

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	Elton    MasterConfig   `toml:"elton"`
	Masters  []MasterConfig `toml:"master"`
	Slave    SlaveConfig    `toml:"slave"`
	Database DBConfig       `toml:"database"`
}

type MasterConfig struct {
	Name     string `toml:"name"`
	HostName string `toml:"hostname"`
}

type SlaveConfig struct {
	MasterHostName string `toml:"master_hostname"`
	IP             string `toml:"ip"`
	Port           uint64 `toml:"port"`
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
