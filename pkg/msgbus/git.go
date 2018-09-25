package msgbus

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/lfkeitel/yobot/pkg/utils"
)

func init() {
	RegisterMsgBus("git", handleGit)
}

type gitEvent struct {
	Secret     string
	Ref        string
	Commits    []*gitEventCommit
	Repository gitEventRepo
}

type gitEventCommit struct {
	Message   string
	URL       string
	Committer struct {
		Name     string
		Email    string
		Username string
	}
}

type gitEventRepo struct {
	Name     string
	FullName string `json:"full_name"`
	HTMLurl  string
}

func handleGit(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var event gitEvent
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&event); err != nil {
		fmt.Printf("Error unmarshalling git event: %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	conf := GetCtxConfig(ctx)
	routeID := GetCtxRouteID(ctx)

	secret := utils.FirstString(conf.Routes[routeID].Settings["secret"], conf.Routes["git"].Settings["secret"])
	if secret != event.Secret {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	for _, commit := range event.Commits {
		msg := fmt.Sprintf("### Git\n\n**%s** committed to **%s** on branch %s - **%s** - %s",
			commit.Committer.Name,
			event.Repository.FullName,
			event.Ref,
			utils.FirstLine(commit.Message),
			commit.URL,
		)

		DispatchMessage(ctx, msg)
	}
}
