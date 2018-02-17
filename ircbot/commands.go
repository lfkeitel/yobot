package ircbot

import (
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	irc "github.com/lfkeitel/goirc/client"
	"github.com/lfkeitel/yobot/ircchalk"
	"github.com/lfkeitel/yobot/utils"
)

func init() {
	RegisterCommand(".ping", pingCommand)
	RegisterCommand(".help", helpCommand)
	RegisterCommand(".hello", helloCommand)

	RegisterCommand("#ping", ping2Command)
}

var (
	pingCommand = &Command{
		Handler: func(conn *irc.Conn, event *Event) error {
			conn.Privmsg(event.Source, "pong")
			return nil
		},
		Help: "Let's play ping pong",
	}
	helpCommand = &Command{
		Handler: func(conn *irc.Conn, event *Event) error {
			conn.Noticef(event.Source, "%s available commands:", ircchalk.Bold(event.Config.IRC.Nick))

			cmdNames := make([]string, 0, len(commands))
			for name, command := range commands {
				if command != nil {
					cmdNames = append(cmdNames, name)
				}
			}

			sort.Strings(cmdNames)

			var buf strings.Builder
			tabs := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', 0)
			for _, name := range cmdNames {
				tabs.Write([]byte(fmt.Sprintf("   %s\t  %s\n", strings.ToUpper(name), utils.FirstLine(commands[name].Help))))
			}
			tabs.Flush()

			for _, line := range strings.Split(buf.String(), "\n") {
				conn.Notice(event.Source, line)
			}

			return nil
		},
		Help: "List help information",
	}
	helloCommand = &Command{
		Handler: func(conn *irc.Conn, event *Event) error {
			conn.Privmsg(event.Source, "Hello, how are you?")
			return nil
		},
		Help: "Say hello",
	}
	ping2Command = &Command{
		Handler: func(conn *irc.Conn, event *Event) error {
			conn.Privmsg(event.Source, "pong2")
			return nil
		},
		Help: "Let's play double ping pong",
	}
)
