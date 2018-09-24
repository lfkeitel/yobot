package ircbot

import (
	irc "github.com/lfkeitel/goirc/client"
	"github.com/lfkeitel/yobot/config"
)

// taps maps channels or nicks to named handlers
var taps = map[string]map[string]CommandHandler{}

// RegisterTap will register a handler to receive all PRIVMSG and ACTION messages
// from the IRC server. Any messages sent by the bot will also be dispatched.
// name should be a unique name to identify the handler. t can be any number
// of channels or nicknames. This function will not join any channels in t
// if it's not already joined.
func RegisterTap(handler CommandHandler, name string, t ...string) {
	for _, target := range t {
		if taps[target] == nil {
			taps[target] = make(map[string]CommandHandler)
		}
		taps[target][name] = handler
	}
}

// UnregisterTap removes a handler from the bot wiretap. name must be the same
// as used in RegisterTap. t can be any number of channels or nicknames.
func UnregisterTap(name string, t ...string) {
	if len(t) == 0 {
		for _, handlers := range taps {
			delete(handlers, name)
		}
	} else {
		for _, target := range t {
			if taps[target] == nil {
				continue
			}
			delete(taps[target], name)
		}
	}
}

func registerTapHandlers(conn *irc.Conn, conf *config.Config) {
	tapHandler := func(conn *irc.Conn, line *irc.Line) {
		event := &Event{
			Line:    line,
			Source:  line.Target(),
			Args:    line.Args[1:],
			Config:  conf,
			Command: "",
		}

		dispatchTaps(wrapConn(conn, conf), event)
	}

	conn.HandleFunc(irc.ACTION, tapHandler)
	conn.HandleFunc(irc.PRIVMSG, tapHandler)
}

func dispatchTaps(conn *Conn, event *Event) {
	for _, tap := range taps[event.Source] {
		tap(conn, event)
	}
}
