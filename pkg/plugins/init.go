package plugins

import (
	"github.com/lfkeitel/yobot/pkg/bot"
	"github.com/lfkeitel/yobot/pkg/config"
)

type (
	InitFunc     func(conf *config.Config, bot *bot.Bot)
	ShutdownFunc func()
)

var (
	inits     = []InitFunc{}
	shutdowns = []ShutdownFunc{}
	ran       = false
)

func RegisterInit(init InitFunc) {
	inits = append(inits, init)
}

func RegisterShutdown(sd ShutdownFunc) {
	shutdowns = append(shutdowns, sd)
}

func Init(conf *config.Config, bot *bot.Bot) {
	if ran {
		return
	}
	for _, init := range inits {
		init(conf, bot)
	}
}

func Shutdown() {
	for _, sd := range shutdowns {
		sd()
	}
}
