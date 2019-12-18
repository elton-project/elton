package main

import (
	"gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/subsystems"
	"gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/subsystems/controller/simple"
	localStorage "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/subsystems/storage/local"
)

func NewServer(role string) subsystems.Server {
	switch role {
	case "controller":
		return simple.NewServer()
	case "storage":
		return localStorage.NewLocalStorageServer()
	default:
		return nil
	}
}
