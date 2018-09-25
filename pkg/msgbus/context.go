package msgbus

import (
	"context"

	"github.com/lfkeitel/yobot/pkg/config"
)

type contextKey string

// Keys used for context items
const (
	configKey contextKey = "config"
	routeKey  contextKey = "route"
	ircKey    contextKey = "irc"
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
