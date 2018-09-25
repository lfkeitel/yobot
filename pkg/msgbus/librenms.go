package msgbus

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

func init() {
	RegisterMsgBus("librenms", handleLibreNMS)
}

func handleLibreNMS(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	severity := strings.ToUpper(r.Form.Get("severity"))

	if strings.Contains(r.Form.Get("title"), "recovered") {
		severity = "RECOVERY"
	}

	switch severity {
	case "CRITICAL":
		severity = ":bangbang: " + severity
	case "WARNING":
		severity = ":heavy_exclamation_mark: " + severity
	case "RECOVERY":
		severity = ":white_check_mark: " + severity
	}

	msg := fmt.Sprintf("### LibreNMS\n\n**%s** - %s on host %s - %s @ %s",
		severity,
		r.Form.Get("title"),
		r.Form.Get("host"),
		r.Form.Get("rule"),
		r.Form.Get("timestamp"),
	)

	DispatchMessage(ctx, msg)
	w.Write([]byte(`{"accepted": true}`))
}
