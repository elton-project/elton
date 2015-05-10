package elton

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	Elton    EltonConfig  `toml:"elton"`
	Backup   BackupConfig `toml:"backup"`
	Database DBConfig     `toml:"database"`
}

type EltonConfig struct {
	HostName string `toml:"hostname"`
	Port     string `toml:"port"`
	Dir      string `toml:"dir"`
}

type DBConfig struct {
	User     string `toml:"user"`
	Pass     string `toml:"pass"`
	HostName string `toml:"hostname"`
	Port     string `toml:"port"`
	DBName   string `toml:"dbname"`
}

type BackupConfig struct {
	HostName string `toml:"hostname"`
	Port     string `toml:"port"`
}

func Load(path string) (Config, error) {
	var conf Config
	_, err := toml.DecodeFile(path, &conf)
	if err != nil {
		return Config{}, err
	}
	return conf, nil
}
