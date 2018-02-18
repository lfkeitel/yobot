package ircbot

import (
	"fmt"
	"time"

	irc "github.com/lfkeitel/goirc/client"
	"github.com/lfkeitel/yobot/config"
)

// Conn wraps an IRC client connection to provide wiretapping and other functions.
type Conn struct {
	*irc.Conn
	conf *config.Config
}

func wrapConn(c *irc.Conn, conf *config.Config) *Conn {
	return &Conn{
		Conn: c,
		conf: conf,
	}
}

func (conn *Conn) wiretap(t, msg string) {
	line := &irc.Line{
		Nick:  conn.conf.IRC.Nick,
		Ident: conn.conf.IRC.Nick,
		Host:  "ircbot",
		Src:   fmt.Sprintf("%s!%s@%s", conn.conf.IRC.Nick, conn.conf.IRC.Nick, "ircbot"),
		Cmd:   irc.ACTION,
		Args:  []string{t, msg},
		Time:  time.Now(),
	}

	event := &Event{
		Line:   line,
		Source: line.Target(),
		Args:   line.Args[1:],
		Config: conn.conf,
	}

	dispatchTaps(conn, event)
}

// Action sends a CTCP ACTION message "/me". This function will wiretap the message
// to any subscribed handlers.
func (conn *Conn) Action(t, msg string) {
	conn.wiretap(t, msg)
	conn.Ctcp(t, irc.ACTION, msg)
}

// Privmsg sends a message. t can be either a channel or user nick.
// This function will wiretap the message to any subscribed handlers.
func (conn *Conn) Privmsg(t, msg string) {
	conn.wiretap(t, msg)
	conn.Conn.Privmsg(t, msg)
}

// Privmsgln sends a message formatted with Sprintln. The line ending is stripped.
// t can be either a channel or user nick.
// This function will wiretap the message to any subscribed handlers.
func (conn *Conn) Privmsgln(t string, a ...interface{}) {
	msg := fmt.Sprintln(a...)
	// trimming the new-line character added by the fmt.Sprintln function,
	// since it's irrelevant.
	msg = msg[:len(msg)-1]
	conn.Privmsg(t, msg)
}

// Privmsgf sends a message formatted with Sprintf.
// t can be either a channel or user nick.
// This function will wiretap the message to any subscribed handlers.
func (conn *Conn) Privmsgf(t, format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	conn.Privmsg(t, msg)
}
