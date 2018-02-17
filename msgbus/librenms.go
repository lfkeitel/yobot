package msgbus

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/lfkeitel/yobot/ircchalk"
)

func init() {
	RegisterMsgBus("librenms", handleLibreNMS)
}

func handleLibreNMS(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	severity := strings.ToUpper(r.Form.Get("severity"))

	switch severity {
	case "CRITICAL":
		severity = ircchalk.Chalk(ircchalk.Red, "", severity)
	case "WARNING":
		severity = ircchalk.Chalk(ircchalk.Orange, "", severity)
	}
	severity = ircchalk.Bold(severity)

	msg := fmt.Sprintf("LibreNMS: %s - %s on host %s - %s @ %s",
		severity,
		r.Form.Get("title"),
		r.Form.Get("host"),
		r.Form.Get("rule"),
		r.Form.Get("timestamp"),
	)

	DispatchIRCMessage(ctx, msg)
	w.Write([]byte(`{"accepted": true}`))
}
