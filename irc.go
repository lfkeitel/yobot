package main

import (
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	irc "github.com/lfkeitel/goirc/client"
	"github.com/lfkeitel/goirc/logging"
)

func startIRCBot(conf *config, quit chan bool) {
	if conf.Main.ExtraDebug {
		logging.SetLogger(&logging.StdoutLogger{})
	}

	if conf.IRC.SASL.UseSASL && conf.IRC.SASL.Login == "" {
		conf.IRC.SASL.Login = conf.IRC.Nick
	}

	if conf.IRC.SASL.UseSASL {
		conf.IRC.TLS = true
	}

	cfg := irc.NewConfig(conf.IRC.Nick)
	cfg.SSL = conf.IRC.TLS
	cfg.SSLConfig = &tls.Config{InsecureSkipVerify: conf.IRC.InsecureTLS}
	cfg.Server = fmt.Sprintf("%s:%d", conf.IRC.Server, conf.IRC.Port)
	cfg.NewNick = func(n string) string { return n + "^" }
	cfg.Me.Ident = conf.IRC.Nick
	cfg.Flood = true
	cfg.SplitLen = 2000
	cfg.Version = "Yobot v1"

	cfg.UseSASL = conf.IRC.SASL.UseSASL
	cfg.SASLLogin = conf.IRC.SASL.Login
	cfg.SASLPassword = conf.IRC.SASL.Password
	c := irc.Client(cfg)

	c.HandleFunc(irc.CONNECTED, func(conn *irc.Conn, line *irc.Line) {
		fmt.Println("Connected to IRC server, joining channels")

		channels := make(map[string]bool)
		for _, channel := range conf.IRC.Channels {
			channels[channel] = true
		}

		if conf.IRC.AutoJoinAlertChannels {
			for _, route := range conf.Routes {
				for _, channel := range route.Channels {
					channels[channel] = true
				}
			}
		}

		for channel := range channels {
			if channel[0] == '#' {
				fmt.Printf("Joining %s\n", channel)
				conn.Join(channel)
			}
		}

		ircConn = conn
	})

	c.HandleFunc(irc.DISCONNECTED, func(conn *irc.Conn, line *irc.Line) {
		<-time.After(2 * time.Second)
		if err := conn.Connect(); err != nil {
			fmt.Printf("Error attempting reconnection: %s\n", err)
			close(quit)
		}
	})

	c.HandleFunc(irc.PRIVMSG, func(conn *irc.Conn, line *irc.Line) {
		if conf.Main.Debug {
			fmt.Printf("%#v\n", line)
		}

		recipient := line.Args[0]
		if recipient == conf.IRC.Nick {
			recipient = line.Nick
		}

		fmt.Println(line.Args[1])
		if isIRCChannel(recipient) &&
			line.Args[1][0] != '.' &&
			!strings.HasPrefix(line.Args[1], conf.IRC.Nick+", ") {
			return
		}
		conn.Privmsg(recipient, "All I do is relay messages.")
	})

	if err := c.Connect(); err != nil {
		fmt.Printf("Connection error: %s\n", err.Error())
	}

	<-quit
	c.Close()
}

func parseCommandLine(line string) []string {
	return strings.Split(line, " ")
}

func isIRCChannel(name string) bool {
	return len(name) > 0 && name[0] == '#'
}
