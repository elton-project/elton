package main

import (
	"github.com/kelseyhightower/envconfig"
	"net"
	"strings"
)

type EnvConfig struct {
	ControllerListenAddr tcpAddrDecoder `split_words:"true"`
	Controllers          tcpAddrsDecoder
}

func loadFromEnvironment() (*EnvConfig, error) {
	conf := &EnvConfig{}
	err := envconfig.Process("ELTON", conf)
	if err != nil {
		return nil, err
	}
	return conf, err
}

type tcpAddrDecoder struct {
	Addr net.Addr
}

func (addr *tcpAddrDecoder) Decode(value string) (err error) {
	addr.Addr, err = net.ResolveTCPAddr("tcp", value)
	if err != nil {
		addr.Addr = nil
	}
	return
}

type tcpAddrsDecoder struct {
	Addrs []net.Addr
}

func (addrs *tcpAddrsDecoder) Decode(value string) error {
	for _, s := range strings.Split(value, ",") {
		var decoder tcpAddrDecoder
		if err := decoder.Decode(s); err != nil {
			addrs.Addrs = nil
			return err
		}
		addrs.Addrs = append(addrs.Addrs, decoder.Addr)
	}
	return nil
}
