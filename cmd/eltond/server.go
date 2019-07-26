package main

import (
	"gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/subsystems"
	localStorage "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/subsystems/storage/local"
)

func NewServer(role string) subsystems.Server {
	switch role {
	case "controller":
		panic("todo")
	case "storage":
		return localStorage.NewLocalStorageServer()
	default:
		return nil
	}
}
