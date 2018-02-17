package msgbus

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"

	irc "github.com/lfkeitel/goirc/client"

	"github.com/lfkeitel/yobot/config"
	"github.com/lfkeitel/yobot/utils"
)

type BusHandler func(context.Context, http.ResponseWriter, *http.Request)
type MuxHandler func(*config.Config) http.HandlerFunc

var (
	handlers    = map[string]BusHandler{}
	handlerLock sync.Mutex

	muxHandlers = map[string]MuxHandler{
		"/msgbus/": msgbusHandler,
	}
	muxHandlersLock sync.Mutex
)

func RegisterMsgBus(id string, handler BusHandler) {
	handlerLock.Lock()
	defer handlerLock.Unlock()
	if _, exists := handlers[id]; exists {
		panic(fmt.Sprintf("handler id %s is already registered", id))
	}
	handlers[id] = handler
}

func RegisterMuxHandler(path string, handler MuxHandler) {
	muxHandlersLock.Lock()
	defer muxHandlersLock.Unlock()
	if _, exists := muxHandlers[path]; exists {
		panic(fmt.Sprintf("handler path %s is already registered", path))
	}
	muxHandlers[path] = handler
}

type ContextKey string

// Keys used for context items
const (
	ConfigKey ContextKey = "config"
	RouteKey  ContextKey = "route"
	IRCKey    ContextKey = "irc"
)

var ircConn *irc.Conn

func SetIRCConn(c *irc.Conn) {
	ircConn = c
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

		ctx := context.WithValue(context.Background(), RouteKey, routeID)
		ctx = context.WithValue(ctx, ConfigKey, conf)
		ctx = context.WithValue(ctx, IRCKey, ircConn)

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

func DispatchIRCMessage(conf *config.Config, source, f string, a ...interface{}) {
	channels := conf.Routes[source].Channels
	if len(channels) == 0 {
		channels = conf.Routes["default"].Channels
	}

	for _, channel := range channels {
		ircConn.Privmsgf(channel, f, a...)
	}
}
