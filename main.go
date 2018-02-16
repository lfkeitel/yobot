package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	irc "github.com/lfkeitel/goirc/client"
)

var (
	configFile string

	ircConn *irc.Conn
)

func init() {
	flag.StringVar(&configFile, "c", "config.toml", "Config file")

	rand.Seed(time.Now().UnixNano())
}

func main() {
	flag.Parse()

	conf, err := loadConfig(configFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if conf.Main.ExtraDebug {
		conf.Main.Debug = true
	}

	quit := make(chan bool)
	go startIRCBot(conf, quit)
	go startHTTPServer(conf, quit)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	<-shutdown
	fmt.Println("Stopping")
	close(quit)
	time.Sleep(5)
}

func firstString(o ...string) string {
	for _, s := range o {
		if s != "" {
			return s
		}
	}

	return ""
}

func firstLine(s string) string {
	return strings.Split(s, "\n")[0]
}
