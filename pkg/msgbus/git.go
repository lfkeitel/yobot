package msgbus

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/lfkeitel/yobot/pkg/utils"
)

const gitPostEmoji = ":large_blue_circle:"

func init() {
	RegisterMsgBus("git", handleGit)
}

type gitEvent struct {
	Secret     string // Gitea only
	Ref        string
	Commits    []*gitEventCommit
	Repository gitEventRepo
}

type gitEventCommit struct {
	Message string
	URL     string
	Author  struct { // GitHub only
		Name  string
		Email string
	}
	Committer struct { // Gitea only
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
	if r.Header.Get("X-Hub-Signature") != "" { // Gitea doesn't send this header
		handleGitHubEvent(ctx, w, r)
		return
	}
	handleGiteaEvent(ctx, w, r)
}

func handleGiteaEvent(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var event gitEvent
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&event); err != nil {
		fmt.Printf("Error unmarshalling git event: %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	secret := RouteSetting(ctx, "secret").(string)

	if secret != event.Secret {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	for _, commit := range event.Commits {
		msg := fmt.Sprintf("### Git\n\n%s **%s** committed to **%s** on branch %s - **%s** - %s",
			gitPostEmoji,
			commit.Committer.Name,
			event.Repository.FullName,
			event.Ref,
			utils.FirstLine(commit.Message),
			commit.URL,
		)

		DispatchMessage(ctx, msg)
	}
}

func handleGitHubEvent(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}

	secret := RouteSetting(ctx, "secret").(string)

	if err := checkGitHubSig(secret, r.Header.Get("X-Hub-Signature"), body); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var event gitEvent
	if err := json.Unmarshal(body, &event); err != nil {
		fmt.Printf("Error unmarshalling git event: %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for _, commit := range event.Commits {
		msg := fmt.Sprintf("### Git\n\n%s **%s** committed to **%s** on branch %s - **%s** - %s",
			gitPostEmoji,
			commit.Author.Name,
			event.Repository.FullName,
			event.Ref,
			utils.FirstLine(commit.Message),
			commit.URL,
		)

		DispatchMessage(ctx, msg)
	}
}

func checkGitHubSig(secret, header string, body []byte) error {
	if !strings.HasPrefix(header, "sha1=") {
		return errors.New("No signature header from GitHub event")
	}
	header = header[5:]
	hexbytes, err := hex.DecodeString(header)
	if err != nil {
		return errors.New("Malformatted signature header from GitHub event")
	}

	if !checkGitHubHash(secret, body, hexbytes) {
		return errors.New("Invalid signature hash from GitHub event")
	}
	return nil
}

func checkGitHubHash(key string, message, expectedMAC []byte) bool {
	mac := hmac.New(sha1.New, []byte(key))
	mac.Write(message)
	messageMAC := mac.Sum(nil)
	return hmac.Equal(messageMAC, expectedMAC)
}
