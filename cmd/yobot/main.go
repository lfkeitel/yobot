package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lfkeitel/yobot/pkg/bot"
	"github.com/lfkeitel/yobot/pkg/config"
	"github.com/lfkeitel/yobot/pkg/msgbus"
	"github.com/lfkeitel/yobot/pkg/plugins"
	"github.com/lfkeitel/yobot/pkg/utils"
)

var (
	configFile     string
	testPluginFlag bool

	debug       bool
	extraDebug  bool
	versionInfo bool

	version   = ""
	buildTime = ""
	builder   = ""
	goversion = ""
)

func init() {
	flag.StringVar(&configFile, "c", "config.toml", "Config file")
	flag.BoolVar(&debug, "debug", false, "Debug mode")
	flag.BoolVar(&extraDebug, "debug2", false, "Extra debug mode")
	flag.BoolVar(&testPluginFlag, "tp", false, "Test loading plugins")
	flag.BoolVar(&versionInfo, "v", false, "Print version information")

	rand.Seed(time.Now().UnixNano())
}

func main() {
	flag.Parse()

	if versionInfo {
		displayVersionInfo()
		return
	}

	conf, err := config.LoadConfig(configFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if extraDebug {
		conf.Main.ExtraDebug = extraDebug
	}
	if debug {
		conf.Main.Debug = debug
	}

	if conf.Main.ExtraDebug {
		conf.Main.Debug = true
	}

	if err := os.MkdirAll(conf.Main.DataDir, 0755); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := plugins.Load(conf.Main.ModulesDir, conf.Main.Modules); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if testPluginFlag {
		return
	}

	quit := utils.GetQuitChan()
	done := make(chan bool, 2)
	if err := bot.Start(conf, quit, done); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := msgbus.Start(conf, quit, done); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	plugins.Init(conf, bot.GetBot())

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	select {
	case <-quit:
	case <-shutdown:
		close(quit)
	}

	fmt.Println("Stopping")
	plugins.Shutdown()

	timer := time.NewTimer(5 * time.Second)

	// Wait for all services to stop
	for i := 0; i < 2; i++ {
		select {
		case <-done:
		case <-timer.C:
			fmt.Println("Services didn't stop fast enough")
			os.Exit(1)
		}
	}
}

func displayVersionInfo() {
	fmt.Printf(`Yobot

Version:     %s
Built:       %s
Compiled by: %s
Go version:  %s
`, version, buildTime, builder, goversion)
}
