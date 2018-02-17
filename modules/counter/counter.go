package main

import (
	"context"
	"net/http"

	"github.com/lfkeitel/yobot/msgbus"
)

func init() {
	msgbus.RegisterMsgBus("counter", handleCounter)
}

var counter = 0

func handleCounter(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	counter++
	msgbus.DispatchIRCMessage(ctx, "Counter: %d", counter)
}

func main() {}
