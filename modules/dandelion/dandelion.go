package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/lfkeitel/yobot/pkg/bot"
	"github.com/lfkeitel/yobot/pkg/plugins"

	"github.com/lfkeitel/yobot/pkg/config"
	"github.com/lfkeitel/yobot/pkg/utils"
)

type dandelionConfig struct {
	URL, ApiKey string
	Channels    []string
}

var (
	dconf  *dandelionConfig
	lastID = 0
)

func init() {
	plugins.RegisterInit(dandelionInit)
}

func dandelionInit(conf *config.Config, bot *bot.Bot) {
	var dc dandelionConfig
	if err := utils.FillStruct(&dc, conf.Modules["dandelion"]); err != nil {
		panic(err)
	}
	dconf = &dc

	go startDandelionCheck(bot)
}

type dandelionResp struct {
	Status      string
	Errorcode   int
	Module      string
	RequestTime string
	Data        struct {
		Logs     []dandelionLog
		Metadata dandelionMetadata
	}
}

type dandelionLog struct {
	ID          int
	DateCreated string
	TimeCreated string
	Title       string
	Body        string
	UserID      int
	Category    string
	IsEdited    bool
	Fullname    string
	CanEdit     bool
}

type dandelionMetadata struct {
	Limit       int
	LogSize     int
	Offset      int
	ResultCount int
}

func startDandelionCheck(bot *bot.Bot) {
	readAPI := dconf.URL + "/api/logs/read"
	params := make(url.Values)
	params.Set("apikey", dconf.ApiKey)
	params.Set("limit", "10")
	readAPI = readAPI + "?" + params.Encode()

	for {
		var decoder *json.Decoder
		var apiResp dandelionResp
		var newID int

		resp, err := http.Get(readAPI)
		if err != nil {
			fmt.Println(err)
			goto sleep
		}
		defer resp.Body.Close()

		decoder = json.NewDecoder(resp.Body)
		if err := decoder.Decode(&apiResp); err != nil {
			fmt.Println(err)
			goto sleep
		}

		// Bad API request
		if apiResp.Errorcode != 0 {
			fmt.Println(apiResp.Status)
			goto sleep
		}

		// No returned logs
		if apiResp.Data.Metadata.ResultCount == 0 {
			goto sleep
		}

		newID = apiResp.Data.Logs[0].ID
		if lastID == 0 {
			lastID = newID
			goto sleep
		}
		if newID <= lastID {
			goto sleep
		}

		for _, log := range apiResp.Data.Logs {
			if log.ID > lastID {
				msg := fmt.Sprintf("### Dandelion\n\n**%s** (%s) <%s/log/%d>", log.Title, log.Fullname, dconf.URL, log.ID)

				for _, channel := range dconf.Channels {
					bot.SendMsgTeamChannel(channel, msg)
				}
			}
		}
		lastID = newID

	sleep:
		time.Sleep(10 * time.Second)
	}
}

func main() {}
