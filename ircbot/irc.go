package ircbot

import (
	"crypto/tls"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	irc "github.com/lfkeitel/goirc/client"
	"github.com/lfkeitel/goirc/logging"

	"github.com/lfkeitel/yobot/config"
)

var ircConn *irc.Conn

func GetBot() *irc.Conn {
	return ircConn
}

type Event struct {
	*irc.Line
	Source string
	Args   []string
	Config *config.Config
}

type CommandHandler func(conn *irc.Conn, event *Event) error
type Command struct {
	name    string
	Handler CommandHandler
	Help    string
}

var (
	commands     = map[string]*Command{}
	commandsLock sync.Mutex
)

func RegisterCommand(cmd string, command Command) {
	commandsLock.Lock()
	defer commandsLock.Unlock()
	if _, exists := commands[cmd]; exists {
		panic(fmt.Sprintf("IRC command %s is already registered", cmd))
	}
	commands[cmd] = &command
}

func Start(conf *config.Config, quit, done chan bool) error {
	ready := make(chan bool)
	go start(conf, quit, done, ready)

	select {
	case <-ready:
	case <-time.After(5 * time.Second):
		return errors.New("Failed connecting to IRC server")
	}
	return nil
}

func start(conf *config.Config, quit, done, ready chan bool) {
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
	cfg.Me.Name = strings.Title(conf.IRC.Nick)
	cfg.Flood = true
	cfg.SplitLen = 2000
	cfg.Version = "Yobot v1"

	cfg.UseSASL = conf.IRC.SASL.UseSASL
	cfg.SASLLogin = conf.IRC.SASL.Login
	cfg.SASLPassword = conf.IRC.SASL.Password
	c := irc.Client(cfg)

	var chans []string

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
				chans = append(chans, channel)
			}
		}

		ircConn = conn
		close(ready)
	})

	closing := false
	c.HandleFunc(irc.DISCONNECTED, func(conn *irc.Conn, line *irc.Line) {
		if closing {
			return
		}

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

		if conf.Main.Debug {
			fmt.Println(line.Args[1])
		}
		if IsChannel(recipient) &&
			!isCommand(line.Args[1]) &&
			!strings.HasPrefix(line.Args[1], conf.IRC.Nick+", ") {
			return
		}

		addressedByName := false
		args := parseCommandLine(line.Args[1])
		if args[0] == conf.IRC.Nick+"," {
			args = args[1:]
			addressedByName = true
		}

		if len(args) == 0 {
			return
		}

		if (addressedByName || !IsChannel(recipient)) && !isCommand(args[0]) {
			args[0] = "." + args[0]
		}

		if conf.Main.Debug {
			fmt.Printf("Source: %s, Line: %v\n", recipient, args)
		}

		cmd := strings.ToLower(args[0])
		args = args[1:]

		var handler CommandHandler
		for name, chandler := range commands {
			if name == cmd {
				handler = chandler.Handler
				break
			}
		}

		if handler == nil {
			if IsChannel(recipient) {
				recipient = line.Nick
			}

			conn.Privmsg(recipient, "Please try .help")
			return
		}

		event := &Event{
			Line:   line,
			Source: recipient,
			Args:   args,
			Config: conf,
		}

		if err := handler(conn, event); err != nil {
			fmt.Println(err)
		}
	})

	if err := c.Connect(); err != nil {
		fmt.Printf("Connection error: %s\n", err.Error())
	}

	<-quit
	closing = true
	fmt.Println("Disconnecting from IRC server")

	for _, channel := range chans {
		fmt.Printf("Leaving %s\n", channel)
		c.Part(channel, "Bye, bye")
	}
	c.Quit("Bye everyone!")

	<-time.After(1 * time.Second) // Give messages time to send
	c.Close()
	fmt.Println("Disconnected for IRC server")
	done <- true
}

func parseCommandLine(line string) []string {
	return strings.Split(line, " ")
}

func IsChannel(name string) bool {
	return len(name) > 0 && name[0] == '#'
}

func isCommand(s string) bool {
	if s == "" {
		return false
	}
	return s[0] == '.' || s[0] == '#' || s[0] == '!'
}
