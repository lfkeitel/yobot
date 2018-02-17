package msgbus

import (
	"context"
	"fmt"
	"net/http"

	"github.com/lfkeitel/yobot/config"
)

func init() {
	RegisterMsgBus("librenms", handleLibreNMS)
}

func handleLibreNMS(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	conf := ctx.Value(ConfigKey).(*config.Config)

	r.ParseForm()
	msg := fmt.Sprintf("LibreNMS: %s %s on host %s - %s @ %s",
		r.Form.Get("severity"),
		r.Form.Get("title"),
		r.Form.Get("host"),
		r.Form.Get("rule"),
		r.Form.Get("timestamp"),
	)

	DispatchIRCMessage(conf, ctx.Value(RouteKey).(string), msg)
	w.Write([]byte(`{"accepted": true}`))
}
