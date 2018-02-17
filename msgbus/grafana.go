package msgbus

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/lfkeitel/yobot/config"
)

func init() {
	RegisterMsgBus("grafana", handleGrafana)
}

func handleGrafana(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	conf := ctx.Value(ConfigKey).(*config.Config)
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

	var alert grafanaAlert
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&alert); err != nil {
		fmt.Printf("Error unmarshalling Grafana alert: %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	DispatchIRCMessage(conf, ctx.Value(RouteKey).(string), "%s - %s", alert.Title, alert.Message)
	w.Write([]byte(`{"accepted": true}`))
}
