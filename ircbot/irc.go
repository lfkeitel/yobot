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

// GetBot returns the existing IRC connection.
func GetBot() *irc.Conn {
	return ircConn
}

// An Event is an IRC message with some niceties added.
type Event struct {
	// Line is the underlying IRC library event object
	*irc.Line

	// Source is the channel or nickname where the event originated
	Source string

	// Args is the space-split text after the command
	Args []string

	// Config is the application configuration
	Config *config.Config

	// Command is the command name issued such as ".ping"
	Command string
}

// A CommandHandler actually handles processing a command.
type CommandHandler func(conn *irc.Conn, event *Event) error

// A Command contains information about a command.
type Command struct {
	name string

	// Handler will be called when a message comes in for a given command
	Handler CommandHandler

	// Help is a one line description of the command shown when the user
	// issues ".help"
	Help string
}

var (
	commands     = map[string]*Command{}
	commandsLock sync.Mutex
	defaultCmds  = map[string]*Command{}
)

// RegisterCommand allows a module to register a command text with a handler.
// Commands must begin with a '.', '#', or '!'. If a user is direct messaging
// the bot, a '.' is prepended to any command that doesn't already look like
// a command. E.g., if a user sends "ping" it will be transformed into ".ping"
// before being routed.
func RegisterCommand(cmd string, command *Command) {
	commandsLock.Lock()
	defer commandsLock.Unlock()
	if !isCommand(cmd) {
		panic(fmt.Sprintf("%s is not a valid command", cmd))
	}

	if _, exists := commands[cmd]; exists {
		panic(fmt.Sprintf("IRC command %s is already registered", cmd))
	}
	commands[cmd] = command
}

// SetDefaultCommand allows a module to route any non-existant command messages to
// a particular command. This can be used to engage a user mode where the user doesn't
// need to prepend every message with a command. The module is responsible for clearing
// the default command when it's done. t may be a user nick or channel name.
// Setting a new default command will override any previous default.
func SetDefaultCommand(t string, command *Command) {
	fmt.Printf("Setting default handler for %s\n", t)
	defaultCmds[t] = command
}

// ClearDefaultCommand will remove the default non-existant command command from a target.
// t may be a user nick or channel name.
func ClearDefaultCommand(t string) {
	fmt.Printf("Clearing default handler for %s\n", t)
	defaultCmds[t] = nil
}

// Start will attempt to the start the IRC client.
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
		defer func() {
			if r := recover(); r != nil {
				fmt.Println(r)
			}
		}()

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
				if chandler != nil {
					handler = chandler.Handler
				}
				break
			}
		}

		if handler == nil {
			d := defaultCmds[recipient]
			if d != nil {
				handler = d.Handler
			}
		}

		if handler == nil {
			conn.Privmsg(recipient, "Please try .help")
			return
		}

		event := &Event{
			Line:    line,
			Source:  recipient,
			Args:    args,
			Config:  conf,
			Command: cmd,
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

// IsChannel returns if a string looks like a channel name.
func IsChannel(name string) bool {
	return len(name) > 0 && name[0] == '#'
}

func isCommand(s string) bool {
	if s == "" {
		return false
	}
	return s[0] == '.' || s[0] == '#' || s[0] == '!'
}
