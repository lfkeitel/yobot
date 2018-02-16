package main

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/naoina/toml"
)

type config struct {
	Main   mainConfig
	IRC    ircConfig
	HTTP   httpConfig
	Routes map[string]routeConfig
}

type mainConfig struct {
	Debug      bool
	ExtraDebug bool
}

type ircConfig struct {
	Server                string
	Port                  int
	Nick                  string
	TLS                   bool
	InsecureTLS           bool
	Channels              []string
	AutoJoinAlertChannels bool

	SASL struct {
		UseSASL  bool `toml:"sasl"`
		Login    string
		Password string
	}
}

type httpConfig struct {
	Address string
}

type routeConfig struct {
	Enabled  bool
	Channels []string
	Username string
	Password string
	Alias    string
	Settings map[string]string
}

func loadConfig(filename string) (conf *config, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		}
	}()

	if filename == "" {
		filename = "config.toml"
	}

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	var con config
	if err := toml.Unmarshal(buf, &con); err != nil {
		return nil, err
	}
	return setSensibleDefaults(&con)
}

func setSensibleDefaults(con *config) (*config, error) {
	return con, nil
}
