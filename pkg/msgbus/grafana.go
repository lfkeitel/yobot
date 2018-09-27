package msgbus

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func init() {
	RegisterMsgBus("grafana", handleGrafana)
}

type grafanaAlert struct {
	EvalMatches []struct {
		Value  int
		Metric string
		Tags   []string
	} `json:"evalMatches"`
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

	if strings.HasPrefix(alert.Title, "[OK]") {
		alert.Title = strings.Replace(alert.Title, "[OK]", ":white_check_mark:", 1)
	} else if strings.HasPrefix(alert.Title, "[Alerting]") {
		alert.Title = strings.Replace(alert.Title, "[Alerting]", ":bangbang:", 1)
	} else if strings.HasPrefix(alert.Title, "[No Data]") {
		alert.Title = strings.Replace(alert.Title, "[No Data]", ":heavy_exclamation_mark:", 1)
	}

	DispatchMessage(ctx, "### Grafana\n\n**%s** - %s", alert.Title, alert.Message)
	w.Write([]byte(`{"accepted": true}`))
}
