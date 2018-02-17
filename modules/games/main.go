package main

import (
	"fmt"
	"strings"

	irc "github.com/lfkeitel/goirc/client"
	"github.com/lfkeitel/yobot/ircbot"
)

var defaultGameCmd *ircbot.Command

func init() {
	defaultGameCmd = &ircbot.Command{Handler: processMsg}

	ircbot.RegisterCommand(".play", &ircbot.Command{
		Handler: processMsg,
		Help:    "Start playing a game",
	})
	ircbot.RegisterCommand(".stopplaying", &ircbot.Command{
		Handler: processMsg,
		Help:    "Stop playing a game",
	})
	ircbot.RegisterCommand(".playing", &ircbot.Command{
		Handler: processMsg,
		Help:    "Are you playing a game?",
	})
	ircbot.RegisterCommand(".games", &ircbot.Command{
		Handler: processMsg,
		Help:    "List available games",
	})
}

func main() {}

func processMsg(conn *irc.Conn, event *ircbot.Event) error {
	line := event.Line
	args := event.Args

	if event.Config.Main.Debug {
		fmt.Printf("%s %#v\n", event.Command, args)
	}

	switch event.Command {
	case ".play":
		startGame(conn, line, args)
	case ".stopplaying":
		stopGame(conn, line, args)
	case ".games":
		conn.Privmsg(event.Source, "Available games: guess.")
	case ".playing":
		if hasActiveGame(line.Nick) {
			conn.Privmsgf(event.Source, "You're playing %s.", getGame(line.Nick).id())
		} else {
			conn.Privmsg(event.Source, "You're not playing a game. Start one by saying '.play <game>'.")
		}
	default:
		if hasActiveGame(line.Nick) {
			getGame(line.Nick).play(conn, line, append([]string{event.Command}, args...))
		} else {
			conn.Notice(event.Source, "Try '.help' instead.")
		}
	}
	return nil
}

func startGame(conn *irc.Conn, line *irc.Line, args []string) {
	if hasActiveGame(line.Nick) {
		conn.Notice(line.Nick, "You're already playing a game. Please stop your current game first.")
		return
	}

	if len(args) != 1 {
		conn.Notice(line.Nick, "I need to know what game you want to play.")
		conn.Notice(line.Nick, "Use the 'games' command to see what I have.")
		return
	}

	switch args[0] {
	case guessingGameID:
		setGame(line.Nick, newGuessingGame())
		getGame(line.Nick).start(conn, line)
	default:
		conn.Notice(line.Nick, "Use the 'games' command to see what I have.")
	}
	fmt.Printf("User %s started game %s\n", line.Nick, args[0])
	ircbot.SetDefaultCommand(line.Nick, defaultGameCmd)
}

func stopGame(conn *irc.Conn, line *irc.Line, args []string) {
	if !hasActiveGame(line.Nick) {
		conn.Notice(line.Nick, "You're not playing a game right now.")
		return
	}

	if len(args) == 0 {
		conn.Notice(line.Nick, "Are you sure you want to stop the game? Say '.stop y'.")
		return
	}

	response := strings.ToLower(args[0])
	if response == "y" || response == "yes" {
		getGame(line.Nick).stop(conn, line)
		conn.Notice(line.Nick, "I was just beginning to have fun...")
	}
	ircbot.ClearDefaultCommand(line.Nick)
}
