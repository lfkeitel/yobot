package msgbus

import (
	"context"

	irc "github.com/lfkeitel/goirc/client"
	"github.com/lfkeitel/yobot/config"
)

func GetCtxRouteID(ctx context.Context) string {
	return ctx.Value(routeKey).(string)
}
func SetCtxRouteID(ctx context.Context, route string) context.Context {
	return context.WithValue(ctx, routeKey, route)
}

func GetCtxConfig(ctx context.Context) *config.Config {
	return ctx.Value(configKey).(*config.Config)
}
func SetCtxConfig(ctx context.Context, conf *config.Config) context.Context {
	return context.WithValue(ctx, configKey, conf)
}

func GetCtxIRC(ctx context.Context) *irc.Conn {
	return ctx.Value(ircKey).(*irc.Conn)
}
func SetCtxIRC(ctx context.Context, conn *irc.Conn) context.Context {
	return context.WithValue(ctx, ircKey, conn)
}
