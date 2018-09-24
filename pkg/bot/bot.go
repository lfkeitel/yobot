package bot

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/lfkeitel/yobot/pkg/config"
	"github.com/mattermost/mattermost-server/model"
)

var bot *Bot

type Bot struct {
	c            *model.Client4
	user         *model.User
	debugChannel *model.Channel
}

// Start will attempt to the start the Mattermost client.
func Start(conf *config.Config, quit, done chan bool) error {
	ready := make(chan bool)
	go start(conf, quit, done, ready)

	select {
	case <-ready:
	case <-time.After(5 * time.Second):
		return errors.New("Failed connecting to Mattermost server")
	}
	return nil
}

func start(conf *config.Config, quit, done, ready chan bool) {
	remoteURL, err := url.Parse(conf.Mattermost.Server)
	if err != nil {
		fmt.Println("Invalid URL")
		return
	}

	bot = &Bot{
		c: model.NewAPIv4Client(conf.Mattermost.Server),
	}

	// Check server is available
	props, resp := bot.c.GetOldClientConfig("")
	if resp.Error != nil {
		fmt.Println("There was a problem pinging the Mattermost server.  Are you sure it's running?")
		fmt.Println(resp.Error)
		return
	}
	fmt.Println("Server detected and is running version " + props["Version"])

	// Login
	user, resp := bot.c.Login(conf.Mattermost.Login.Username, conf.Mattermost.Login.Password)
	if resp.Error != nil {
		fmt.Println("There was a problem logging into the Mattermost server.  Are you sure ran the setup steps from the README.md?")
		fmt.Println(resp.Error)
		return
	}
	bot.user = user

	// team:channel
	debugChan := strings.SplitN(conf.Mattermost.DebugChannel, ":", 2)

	// Get debug team info
	team, err := bot.FindTeam(debugChan[0])
	if err != nil {
		fmt.Printf("Failed to find debug team %s: %s\n", debugChan[0], err.Error())
		return
	}

	// Get debug channel info or make one
	channel, err := bot.FindChannel(debugChan[1], team.Id)
	if err != nil {
		fmt.Println("Debug channel not found, attempting to create")
		if channel, err = makeDebugChannel(bot.c, debugChan[1], team.Id); err != nil {
			fmt.Printf("Failed to make debug channel %s: %s\n", debugChan[1], err.Error())
			return
		}
	}
	bot.debugChannel = channel

	bot.debugMsg("_Yobot has **started**_", "")

	remoteURL.Scheme = "wss"
	fmt.Printf("Connecting to websocket %s\n", remoteURL.String())
	webSocketClient, err := model.NewWebSocketClient4(remoteURL.String(), bot.c.AuthToken)
	if err.(*model.AppError) != nil {
		fmt.Printf("We failed to connect to the web socket: %s\n", err.Error())
		return
	}
	webSocketClient.Listen()

	go func() {
		for {
			select {
			case resp := <-webSocketClient.EventChannel:
				bot.handleMsgFromDebuggingChannel(resp)
			}
		}
	}()

	// go bot.sendPing()
	ready <- true

	<-quit
	fmt.Println("Disconnecting from Mattermost server")
	done <- true
}

func makeDebugChannel(client *model.Client4, name, teamID string) (*model.Channel, error) {
	channel := &model.Channel{}
	channel.Name = name
	channel.DisplayName = "Debugging for Yobot"
	channel.Purpose = "This is used as a test channel for logging bot debug messages"
	channel.Type = model.CHANNEL_OPEN
	channel.TeamId = teamID

	c, resp := client.CreateChannel(channel)
	return c, resp.Error
}

func (b *Bot) sendPing() {
	for {
		b.debugMsg("Ping", "")
		time.Sleep(5 * time.Second)
	}
}

func (b *Bot) debugMsg(msg string, replyToID string) {
	post := &model.Post{}
	post.ChannelId = b.debugChannel.Id
	post.Message = msg
	post.RootId = replyToID
	_, resp := b.c.CreatePost(post)
	if resp.Error != nil {
		fmt.Printf("Failed to send debug message: %s\n", resp.Error.Error())
	}
}

func (b *Bot) handleMsgFromDebuggingChannel(event *model.WebSocketEvent) {
	// If this isn't the debugging channel then lets ingore it
	if event.Broadcast.ChannelId != b.debugChannel.Id {
		return
	}

	// Lets only reponded to messaged posted events
	if event.Event != model.WEBSOCKET_EVENT_POSTED {
		return
	}

	post := model.PostFromJson(strings.NewReader(event.Data["post"].(string)))
	if post != nil {
		// ignore my events
		if post.UserId == b.user.Id {
			return
		}

		// if you see any word matching 'alive' then respond
		if matched, _ := regexp.MatchString(`(?:^|\W)alive(?:$|\W)`, post.Message); matched {
			b.debugMsg("Yes I'm running", post.Id)
			return
		}

		// if you see any word matching 'up' then respond
		if matched, _ := regexp.MatchString(`(?:^|\W)up(?:$|\W)`, post.Message); matched {
			b.debugMsg("Yes I'm running", post.Id)
			return
		}

		// if you see any word matching 'running' then respond
		if matched, _ := regexp.MatchString(`(?:^|\W)running(?:$|\W)`, post.Message); matched {
			b.debugMsg("Yes I'm running", post.Id)
			return
		}

		// if you see any word matching 'hello' then respond
		if matched, _ := regexp.MatchString(`(?:^|\W)hello(?:$|\W)`, post.Message); matched {
			b.debugMsg("Yes I'm running", post.Id)
			return
		}

		// if you see any word matching 'hello' then respond
		if matched, _ := regexp.MatchString(`(?:^|\W)ping(?:$|\W)`, post.Message); matched {
			b.debugMsg("pong", post.Id)
			return
		}
	}

	b.debugMsg("I did not understand you!", post.Id)
}
