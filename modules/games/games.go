package main

import (
	"math/rand"
	"strconv"

	irc "github.com/lfkeitel/goirc/client"
	"github.com/lfkeitel/yobot/ircbot"
)

const (
	guessingGameID = "guess"
)

type game interface {
	id() string
	isActive() bool
	start(conn *ircbot.Conn, line *irc.Line)
	stop(conn *ircbot.Conn, line *irc.Line)
	play(conn *ircbot.Conn, line *irc.Line, args []string)
}

var games = make(map[string]game)

func hasActiveGame(nick string) bool {
	return games[nick] != nil && games[nick].isActive()
}

func getGame(nick string) game {
	return games[nick]
}

func setGame(nick string, g game) {
	games[nick] = g
}

type baseGame struct {
	active bool
}

func newBaseGame() *baseGame                                { return &baseGame{} }
func (g *baseGame) isActive() bool                          { return g.active }
func (g *baseGame) start(conn *ircbot.Conn, line *irc.Line) { g.active = true }
func (g *baseGame) stop(conn *ircbot.Conn, line *irc.Line)  { g.active = false }

const guessingTries = 6

type guessingGame struct {
	*baseGame
	number    int
	triesLeft int
}

func newGuessingGame() *guessingGame {
	return &guessingGame{baseGame: newBaseGame()}
}

func (g *guessingGame) id() string {
	return guessingGameID
}

func (g *guessingGame) start(conn *ircbot.Conn, line *irc.Line) {
	g.baseGame.active = true
	g.number = int(rand.Int31n(99)) + 1
	g.triesLeft = guessingTries
	conn.Noticef(line.Nick, "Guess a number between 1-100, you have %d tries", guessingTries)
}

func (g *guessingGame) play(conn *ircbot.Conn, line *irc.Line, args []string) {
	if len(args) != 1 {
		conn.Notice(line.Nick, "Just give me your guess please.")
		return
	}

	guessStr := args[0]
	if guessStr[0] == '.' {
		guessStr = guessStr[1:]
	}

	guess, err := strconv.Atoi(guessStr)
	if err != nil || guess < 1 || guess > 100 {
		conn.Notice(line.Nick, "That's not a number between 1 and 100 now is it...")
		return
	}

	if guess == g.number {
		g.stop(conn, line)
		conn.Noticef(line.Nick, "You got it! The number was %d! You guessed the number in %d tries.", g.number, guessingTries-g.triesLeft+1)
		return
	}

	g.triesLeft--
	if g.triesLeft == 0 {
		conn.Noticef(line.Nick, "You ran out of tries. The number was %d. You were %d off.", g.number, abs(guess-g.number))
		g.stop(conn, line)
		return
	}

	if guess > g.number {
		conn.Noticef(line.Nick, "%d is too high, you have %d tries left", guess, g.triesLeft)
	} else if guess < g.number {
		conn.Noticef(line.Nick, "%d is too low, you have %d tries left", guess, g.triesLeft)
	}
}

func abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}
