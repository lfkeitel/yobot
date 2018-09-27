package msgbus

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func init() {
	RegisterMsgBus("general", handleGeneral)
}

func handleGeneral(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	type genericAlert struct {
		Title   string `json:"title"`
		Message string `json:"message"`
	}

	var alert genericAlert
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&alert); err != nil {
		fmt.Printf("Error unmarshalling Generic alert: %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	DispatchMessage(ctx, "%s - %s", alert.Title, alert.Message)
	w.Write([]byte(`{"accepted": true}`))
}
