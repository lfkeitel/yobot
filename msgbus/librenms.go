package msgbus

import (
	"context"
	"fmt"
	"net/http"
)

func init() {
	RegisterMsgBus("librenms", handleLibreNMS)
}

func handleLibreNMS(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	msg := fmt.Sprintf("LibreNMS: %s %s on host %s - %s @ %s",
		r.Form.Get("severity"),
		r.Form.Get("title"),
		r.Form.Get("host"),
		r.Form.Get("rule"),
		r.Form.Get("timestamp"),
	)

	DispatchIRCMessage(ctx, msg)
	w.Write([]byte(`{"accepted": true}`))
}
