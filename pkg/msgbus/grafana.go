package msgbus

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const (
	grafanaEmojiOK       = ":white_check_mark:"
	grafanaEmojiAlerting = ":bangbang:"
	grafanaEmojiNoData   = ":heavy_exclamation_mark:"
)

func init() {
	RegisterMsgBus("grafana", handleGrafana)
}

type grafanaAlert struct {
	// EvalMatches []struct {
	// 	Value  int
	// 	Metric string
	// 	Tags   map[string]string
	// } `json:"evalMatches"`
	ImageURL string `json:"imageUrl"`
	Message  string `json:"message"`
	RuleID   int    `json:"ruleId"`
	RuleName string `json:"ruleName"`
	RuleURL  string `json:"ruleUrl"`
	State    string `json:"state"`
	Title    string `json:"title"`
}

func handleGrafana(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var alert grafanaAlert
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&alert); err != nil {
		fmt.Printf("Error unmarshalling Grafana alert: %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch alert.State {
	case "ok":
		alert.Title = strings.Replace(alert.Title, "[OK]", grafanaEmojiOK, 1)
	case "alerting":
		alert.Title = strings.Replace(alert.Title, "[Alerting]", grafanaEmojiAlerting, 1)
	case "no_data":
		alert.Title = strings.Replace(alert.Title, "[No Data]", grafanaEmojiNoData, 1)
	}

	DispatchMessage(ctx, "### Grafana\n\n**%s** - %s", alert.Title, alert.Message)
	w.Write([]byte(`{"accepted": true}`))
}
