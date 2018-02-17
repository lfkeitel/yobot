package msgbus

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/lfkeitel/yobot/config"
	"github.com/lfkeitel/yobot/utils"
)

func init() {
	RegisterMsgBus("git", handleGit)
}

type gitEvent struct {
	Secret  string
	Ref     string
	Commits []struct {
		Message   string
		URL       string
		Committer struct {
			Name     string
			Email    string
			Username string
		}
	}
	Repository struct {
		Name     string
		FullName string `json:"full_name"`
		HTMLurl  string
	}
}

func handleGit(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var event gitEvent
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&event); err != nil {
		fmt.Printf("Error unmarshalling git event: %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	conf := ctx.Value(configKey).(*config.Config)
	routeID := ctx.Value(routeKey).(string)

	secret := utils.FirstString(conf.Routes[routeID].Settings["secret"], conf.Routes["git"].Settings["secret"])
	if secret != event.Secret {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	for _, commit := range event.Commits {
		msg := fmt.Sprintf("%s committed to %s on branch %s - %s - %s",
			commit.Committer.Name,
			event.Repository.FullName,
			event.Ref,
			utils.FirstLine(commit.Message),
			commit.URL,
		)

		DispatchIRCMessage(ctx, msg)
	}
}
