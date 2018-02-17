package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lfkeitel/yobot/config"
	"github.com/lfkeitel/yobot/ircbot"
	"github.com/lfkeitel/yobot/msgbus"
)

var (
	configFile string
)

func init() {
	flag.StringVar(&configFile, "c", "config.toml", "Config file")

	rand.Seed(time.Now().UnixNano())
}

func main() {
	flag.Parse()

	conf, err := config.LoadConfig(configFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if conf.Main.ExtraDebug {
		conf.Main.Debug = true
	}

	quit := make(chan bool)
	done := make(chan bool, 2)
	if err := ircbot.Start(conf, quit, done); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	msgbus.SetIRCConn(ircbot.GetIRCConn())
	if err := msgbus.Start(conf, quit, done); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	select {
	case <-quit:
	case <-shutdown:
		close(quit)
	}

	fmt.Println("Stopping")
	timer := time.NewTimer(5 * time.Second)

	// Wait for all services to stop
	for i := 0; i < 2; i++ {
		select {
		case <-done:
		case <-timer.C:
			fmt.Println("Services didn't stop fast enought")
			os.Exit(1)
		}
	}
}
