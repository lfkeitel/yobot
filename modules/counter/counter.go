package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/lfkeitel/yobot/pkg/msgbus"
)

func init() {
	msgbus.RegisterMsgBus("counter", handleCounter)
}

var counter = 0

func handleCounter(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	counter++
	msgbus.DispatchMessage(ctx, fmt.Sprintf("Counter: %d", counter))
}

func main() {}
