package bot

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/lfkeitel/yobot/pkg/config"
	"github.com/mattermost/mattermost-server/model"
)

var (
	bot     *Bot
	appconf *config.Config
)

type Bot struct {
	c            *model.Client4
	user         *model.User
	UserID       string
	debugChannel *model.Channel
	chanCache    map[string]*model.Channel
}

func GetBot() *Bot { return bot }

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
		c:         model.NewAPIv4Client(conf.Mattermost.Server),
		chanCache: make(map[string]*model.Channel),
	}
	appconf = conf

	// Check server is available
	props, resp := bot.c.GetOldClientConfig("")
	if resp.Error != nil {
		fmt.Println("There was a problem pinging the Mattermost server.  Are you sure it's running?")
		fmt.Println(resp.Error)
		return
	}
	fmt.Println("Server detected and is running version " + props["Version"])

	// Login
	if err := bot.login(); err != nil {
		fmt.Println(err)
		return
	}

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

	if remoteURL.Scheme == "https" {
		remoteURL.Scheme = "wss"
	} else {
		remoteURL.Scheme = "ws"
	}

	fmt.Printf("Connecting to websocket %s\n", remoteURL.String())
	webSocketClient, err := model.NewWebSocketClient4(remoteURL.String(), bot.c.AuthToken)
	if err.(*model.AppError) != nil {
		fmt.Printf("We failed to connect to the web socket: %s\n", err.Error())
		return
	}
	webSocketClient.Listen()
	bot.debugMsg("_Yobot is connected to the websocket and responding to requests_", "")

	bot.RegisterEventHandler(bot.handleMsgFromDebuggingChannel, bot.debugChannel.Id, model.WEBSOCKET_EVENT_POSTED)

	go func() {
		for {
			select {
			case resp := <-webSocketClient.EventChannel:
				bot.handleEvents(resp)
			}
		}
	}()

	ready <- true

	<-quit
	bot.debugMsg("_Yobot is **stopping**_", "")
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

func (b *Bot) login() error {
	user, resp := bot.c.Login(appconf.Mattermost.Login.Username, appconf.Mattermost.Login.Password)
	if resp.Error != nil {
		return fmt.Errorf("There was a problem logging into the Mattermost server.  Are you sure ran the setup steps from the README.md?\n%s", resp.Error.Error())
	}

	bot.user = user
	bot.UserID = user.Id // Expose ID for other services
	return nil
}

func (b *Bot) debugMsg(msg, replyID string) {
	b.sendMsg(b.debugChannel.Id, msg, replyID)
}

func (b *Bot) SendMsgTeamChannel(name, msg string) error {
	c, cached := b.chanCache[name]
	var err error
	if !cached {
		c, err = b.FindChannelWithTeam(name)
		if err != nil {
			return err
		}
		b.chanCache[name] = c
	}

	return b.sendMsg(c.Id, msg, "")
}

func (b *Bot) sendMsg(id, msg, replyID string) error {
	post := &model.Post{}
	post.ChannelId = id
	post.Message = msg
	post.RootId = replyID

	_, resp := b.c.CreatePost(post)
	if resp.Error == nil {
		return nil
	}

	if resp.Error.Id == "api.context.session_expired.app_error" {
		if err := b.login(); err != nil {
			return errors.New("session expired and failed to login again, please check credentials")
		}

		return b.sendMsg(id, msg, replyID)
	}

	return fmt.Errorf("failed to send message: %s (%s)", resp.Error.Error(), resp.Error.Id)
}
