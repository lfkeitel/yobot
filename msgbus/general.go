package msgbus

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/lfkeitel/yobot/config"
)

func init() {
	RegisterMsgBus("general", handleGeneral)
}

func handleGeneral(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	conf := ctx.Value(ConfigKey).(*config.Config)
	type genericAlert struct {
		Title      string `json:"title"`
		Message    string `json:"message"`
		TitleColor string `json:"title_color"`
	}

	var alert genericAlert
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&alert); err != nil {
		fmt.Printf("Error unmarshalling Generic alert: %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	DispatchIRCMessage(conf, ctx.Value(RouteKey).(string), "%s - %s", alert.Title, alert.Message)
	w.Write([]byte(`{"accepted": true}`))
}
