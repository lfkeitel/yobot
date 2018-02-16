package client

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

type SASLResult struct {
	Failed bool
	Err    error
}

func (conn *Conn) setupSASLCallbacks(result chan<- *SASLResult) {
	conn.HandleFunc(CAP, func(conn *Conn, line *Line) {
		if len(line.Args) == 3 {
			if line.Args[1] == "LS" {
				if !strings.Contains(line.Args[2], "sasl") {
					result <- &SASLResult{true, errors.New("no SASL capability " + line.Args[2])}
				}
			}
			if line.Args[1] == "ACK" {
				if conn.cfg.SASLMech != "PLAIN" {
					result <- &SASLResult{true, errors.New("only PLAIN is supported")}
				}
				conn.Raw("AUTHENTICATE " + conn.cfg.SASLMech)
			}
		}
	})
	conn.HandleFunc(AUTHENTICATE, func(conn *Conn, line *Line) {
		str := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s\x00%s\x00%s", conn.cfg.SASLLogin, conn.cfg.SASLLogin, conn.cfg.SASLPassword)))
		conn.Raw("AUTHENTICATE " + str)
	})
	conn.HandleFunc("901", func(conn *Conn, line *Line) {
		conn.Raw("CAP END")
		conn.Raw("QUIT")
		result <- &SASLResult{true, errors.New(line.Args[1])}
	})
	conn.HandleFunc("902", func(conn *Conn, line *Line) {
		conn.Raw("CAP END")
		conn.Raw("QUIT")
		result <- &SASLResult{true, errors.New(line.Args[1])}
	})
	conn.HandleFunc("903", func(conn *Conn, line *Line) {
		result <- &SASLResult{false, nil}
	})
	conn.HandleFunc("904", func(conn *Conn, line *Line) {
		conn.Raw("CAP END")
		conn.Raw("QUIT")
		result <- &SASLResult{true, errors.New(line.Args[1])}
	})
}
