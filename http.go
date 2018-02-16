package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

var handlers = map[string]busHandler{
	"general":  handleGeneral,
	"grafana":  handleGrafana,
	"librenms": handleLibreNMS,
	"git":      handleGit,
}

type busHandler func(context.Context, http.ResponseWriter, *http.Request)

type ctxKey string

const (
	configKey ctxKey = "config"
	routeKey  ctxKey = "route"
)

func startHTTPServer(conf *config, quit chan bool) {
	mux := http.NewServeMux()
	mux.HandleFunc("/msgbus/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if strings.Count(r.URL.Path, "/") < 2 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		split := strings.Split(r.URL.Path, "/")

		handlerID := split[2] // Used to look up route handler
		routeID := handlerID  // Used to look up route attributes from config
		fmt.Printf("Handler: %s, Alias: %s\n", handlerID, conf.Routes[handlerID].Alias)

		if conf.Routes[handlerID].Alias != "" {
			handlerID = conf.Routes[handlerID].Alias
		}

		handler := handlers[handlerID]
		if handler == nil || !conf.Routes[routeID].Enabled {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		ctx := context.WithValue(context.Background(), routeKey, routeID)
		ctx = context.WithValue(ctx, configKey, conf)

		username := firstString(conf.Routes[routeID].Username, conf.Routes["default"].Username)
		password := firstString(conf.Routes[routeID].Password, conf.Routes["default"].Password)

		if !authenticateHandler(username, password, r) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		handler(ctx, w, r)
	})

	server := &http.Server{Addr: conf.HTTP.Address, Handler: mux}

	go func() {
		<-quit
		server.Shutdown(context.Background())
	}()

	fmt.Println("Starting HTTP server")
	if err := server.ListenAndServe(); err != nil {
		fmt.Println(err)
	}
	fmt.Println("HTTP server stopped")
}

func authenticateHandler(username, password string, r *http.Request) bool {
	if username == "" || password == "" { // No authentication configured
		return true
	}

	rusername, rpassword, ok := r.BasicAuth()
	if !ok {
		return false
	}
	// This isn't meant to be sophisticated, just something simple
	return (username == rusername) && (password == rpassword)
}

func handleTest(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	all, _ := ioutil.ReadAll(r.Body)
	fmt.Printf("Test output: %s\n", string(all))
	fmt.Printf("Test params: %s\n", r.URL.Query().Encode())
}

func handleGeneral(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	conf := ctx.Value(configKey).(*config)
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

	dispatchIRCMessage(conf, ctx.Value(routeKey).(string), "%s - %s", alert.Title, alert.Message)
	w.Write([]byte(`{"accepted": true}`))
}

func handleGrafana(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	conf := ctx.Value(configKey).(*config)
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

	dispatchIRCMessage(conf, ctx.Value(routeKey).(string), "%s - %s", alert.Title, alert.Message)
	w.Write([]byte(`{"accepted": true}`))
}

func handleLibreNMS(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	conf := ctx.Value(configKey).(*config)

	r.ParseForm()
	msg := fmt.Sprintf("LibreNMS: Alert %s on host %s - %s",
		r.Form.Get("title"),
		r.Form.Get("host"),
		r.Form.Get("rule"))

	dispatchIRCMessage(conf, ctx.Value(routeKey).(string), msg)
	w.Write([]byte(`{"accepted": true}`))
}

func handleGit(ctx context.Context, w http.ResponseWriter, r *http.Request) {
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

	var event gitEvent
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&event); err != nil {
		fmt.Printf("Error unmarshalling git event: %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	conf := ctx.Value(configKey).(*config)
	routeID := ctx.Value(routeKey).(string)

	secret := firstString(conf.Routes[routeID].Settings["secret"], conf.Routes["git"].Settings["secret"])
	if secret != event.Secret {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	for _, commit := range event.Commits {
		msg := fmt.Sprintf("%s committed to %s on branch %s - %s - %s",
			commit.Committer.Name,
			event.Repository.FullName,
			event.Ref,
			firstLine(commit.Message),
			commit.URL,
		)

		dispatchIRCMessage(conf, routeID, msg)
	}
}

func dispatchIRCMessage(conf *config, source, f string, a ...interface{}) {
	channels := conf.Routes[source].Channels
	if len(channels) == 0 {
		channels = conf.Routes["default"].Channels
	}

	for _, channel := range channels {
		ircConn.Privmsgf(channel, f, a...)
	}
}
