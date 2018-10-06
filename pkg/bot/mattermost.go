package bot

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mattermost/mattermost-server/model"
)

func (b *Bot) FindChannelWithTeam(name string) (*model.Channel, error) {
	c := strings.SplitN(name, ":", 2)

	team, err := b.FindTeam(c[0])
	if err != nil {
		return nil, err
	}

	return bot.FindChannel(c[1], team.Id)
}

func (b *Bot) FindTeam(name string) (*model.Team, error) {
	name = strings.Replace(name, " ", "-", -1)
	team, resp := b.c.GetTeamByName(name, "")
	if resp.Error != nil {
		return nil, resp.Error
	}
	return team, nil
}

func (b *Bot) FindChannel(name, teamID string) (*model.Channel, error) {
	name = strings.Replace(name, " ", "-", -1)
	channel, resp := b.c.GetChannelByName(name, teamID, "")
	if resp.Error != nil {
		return nil, resp.Error
	}
	return channel, nil
}

type EventHandler func(event *model.WebSocketEvent)
type eventHandler struct {
	h         EventHandler
	channelId string
}

var eventHandlers = make(map[string][]eventHandler, 5)

func (b *Bot) RegisterEventHandler(h EventHandler, channelId string, eventTypes ...string) {
	for _, t := range eventTypes {
		eventHandlers[t] = append(eventHandlers[t], eventHandler{
			h:         h,
			channelId: channelId,
		})
	}
}

func (b *Bot) handleEvents(event *model.WebSocketEvent) {
	handlers := eventHandlers[event.Event]
	if len(handlers) == 0 {
		return
	}

	for _, h := range handlers {
		if h.channelId == "*" || h.channelId == event.Broadcast.ChannelId {
			h.h(event)
		}
	}
}

func (b *Bot) testDirectMessage(event *model.WebSocketEvent) {
	fmt.Printf("%#v\n", event)
}

func (b *Bot) handleMsgFromDebuggingChannel(event *model.WebSocketEvent) {
	post := model.PostFromJson(strings.NewReader(event.Data["post"].(string)))
	if post == nil {
		b.debugMsg("I did not understand you!", post.Id)
	}

	// ignore my events
	if post.UserId == b.user.Id || post.IsSystemMessage() {
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

	b.debugMsg("I did not understand you!", post.Id)
}
