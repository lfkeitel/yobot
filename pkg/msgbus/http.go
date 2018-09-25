package msgbus

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/lfkeitel/yobot/pkg/bot"
	"github.com/lfkeitel/yobot/pkg/config"
	"github.com/lfkeitel/yobot/pkg/utils"
)

type BusHandler func(context.Context, http.ResponseWriter, *http.Request)
type MuxHandler func(*config.Config) http.HandlerFunc

var (
	busHandlers     = map[string]BusHandler{}
	busHandlersLock sync.Mutex

	muxHandlers = map[string]MuxHandler{
		"/msgbus/": msgbusHandler,
	}
	muxHandlersLock sync.Mutex
)

func RegisterMsgBus(id string, handler BusHandler) {
	busHandlersLock.Lock()
	defer busHandlersLock.Unlock()
	if _, exists := busHandlers[id]; exists {
		panic(fmt.Sprintf("handler id %s is already registered", id))
	}
	busHandlers[id] = handler
}

func RegisterMuxHandler(path string, handler MuxHandler) {
	muxHandlersLock.Lock()
	defer muxHandlersLock.Unlock()
	if _, exists := muxHandlers[path]; exists {
		panic(fmt.Sprintf("handler path %s is already registered", path))
	}
	muxHandlers[path] = handler
}

func Start(conf *config.Config, quit, done chan bool) error {
	ready := make(chan bool)
	go start(conf, quit, done, ready)
	<-ready
	return nil
}

func start(conf *config.Config, quit, done, ready chan bool) {
	mux := http.NewServeMux()
	for path, handler := range muxHandlers {
		mux.HandleFunc(path, handler(conf))
	}

	server := &http.Server{Addr: conf.HTTP.Address, Handler: mux}

	go func() {
		<-quit
		server.Shutdown(context.Background())
		fmt.Println("HTTP server stopped")
		done <- true
	}()

	fmt.Println("Starting HTTP server")
	close(ready)
	if err := server.ListenAndServe(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func msgbusHandler(conf *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		r.ParseForm()
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

		handler := busHandlers[handlerID]
		if handler == nil || !conf.Routes[routeID].Enabled {
			if !conf.Routes[routeID].Enabled {
				fmt.Printf("Handler %s is disabled\n", routeID)
			}
			w.WriteHeader(http.StatusNotFound)
			return
		}

		ctx := SetCtxRouteID(context.Background(), routeID)
		ctx = SetCtxConfig(ctx, conf)

		username := utils.FirstString(conf.Routes[routeID].Username, conf.Routes["default"].Username)
		password := utils.FirstString(conf.Routes[routeID].Password, conf.Routes["default"].Password)

		if !authenticateHandler(username, password, r) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		handler(ctx, w, r)
	}
}

func authenticateHandler(username, password string, r *http.Request) bool {
	if password == "" { // No authentication configured
		return true
	}

	if username == "" || username == "-" { // authkey authentication
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" { // No header, check authkey parameter
			keyParam := r.FormValue("authkey")
			if keyParam == "" {
				return false
			}
			return keyParam == password
		}

		header := strings.SplitN(authHeader, " ", 2)
		if len(header) != 2 {
			return false
		}

		if header[0] != "Authkey" {
			return false
		}
		return header[1] == password
	}

	rusername, rpassword, ok := r.BasicAuth()
	if !ok {
		return false
	}
	// This isn't meant to be sophisticated, just something simple
	return (username == rusername) && (password == rpassword)
}

// TestMsgBusHandler will print the body of a request and the URL parameters
// to standard output for testing input data.
func TestMsgBusHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	all, _ := ioutil.ReadAll(r.Body)
	fmt.Printf("Test output: %s\n", string(all))
	fmt.Printf("Test params: %s\n", r.URL.Query().Encode())
}

// DispatchMessage will send a post to the appropriate channels
// based on the message's source bus. The Context must have route and
// conf key.
func DispatchMessage(ctx context.Context, f string, a ...interface{}) {
	conf := GetCtxConfig(ctx)
	source := GetCtxRouteID(ctx)

	channels := conf.Routes["default"].Channels
	// Channel override means use the route's channel setting exclusively
	if conf.Routes[source].ChannelOverride {
		channels = conf.Routes[source].Channels
	} else {
		for _, channel := range conf.Routes[source].Channels {
			channels = append(channels, channel)
		}
	}

	msg := fmt.Sprintf(f, a...)
	for _, channel := range channels {
		if err := bot.GetBot().SendMsgTeamChannel(channel, msg); err != nil {
			fmt.Println(err)
		}
	}
}
