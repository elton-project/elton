package api

import (
	"bytes"
	"io/ioutil"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/higanworks/envmap"
)

type Config struct {
	Master   MasterConfig   `toml:"master"`
	Slave    SlaveConfig    `toml:"slave"`
	Backup   BackupConfig   `toml:"backup"`
	Masters  []MasterConfig `toml:"masters"`
	Database DBConfig       `toml:"database"`
}

type MasterConfig struct {
	Name string `toml:"name"`
	Port uint64 `toml:"port"`
}

type SlaveConfig struct {
	Name       string `toml:"name"`
	GrpcPort   uint64 `toml:"grpc_port"`
	HttpPort   uint64 `toml:"http_port"`
	MasterName string `toml:"master_name"`
	MasterPort uint64 `toml:"master_port"`
	Dir        string `toml:"dir"`
}

type BackupConfig struct {
	Name string `toml:"name"`
	Port uint64 `toml:"port"`
}

type DBConfig struct {
	DBPath string `toml:"dbpath"`
}

func Load(path string) (conf Config, err error) {
	envs := envmap.All()
	rawdata, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	tmpl := template.Must(template.New("config").Parse(string(rawdata)))

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, envs); err != nil {
		return
	}

	_, err = toml.DecodeReader(&buf, &conf)
	return
}
