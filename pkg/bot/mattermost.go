package bot

import (
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
