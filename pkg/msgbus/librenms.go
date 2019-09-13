package msgbus

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/lfkeitel/yobot/librenms"
	"github.com/lfkeitel/yobot/pkg/bot"
	"github.com/lfkeitel/yobot/pkg/config"
	"github.com/lfkeitel/yobot/pkg/utils"
)

func init() {
	RegisterMsgBus("librenms", handleLibreNMS)
}

var (
	libreNMSClient *librenms.Client

	routeRegexs = make(map[string][]contact, 1)
)

type contact struct {
	match   *regexp.Regexp
	channel string
}

func handleLibreNMS(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	alertTitle := utils.StringOrDefault(r.Form.Get("title"), "%TITLE%")
	alertHost := utils.StringOrDefault(r.Form.Get("host"), "%HOST%")
	alertSysName := utils.StringOrDefault(r.Form.Get("sysName"), "%SYSNAME%")
	alertSeverity := strings.ToUpper(utils.StringOrDefault(r.Form.Get("severity"), "CRITICAL"))
	alertMsg := strings.ToUpper(utils.StringOrDefault(r.Form.Get("message"), "%MESSAGE%"))

	// LibreNMS sends a critical severity for recovered because the alert itself was
	// critical. We use a special severity tag if the alert is a recovery event.
	if strings.Contains(alertTitle, "recovered") {
		alertSeverity = "RECOVERY"
	}

	// Let the client go on its merry way. We have everything we need now.
	w.Write([]byte(`{"accepted": true}`))

	// Add emojis to the alerts for added emphasis
	switch alertSeverity {
	case "CRITICAL":
		alertSeverity = ":bangbang: " + alertSeverity
	case "WARNING":
		alertSeverity = ":heavy_exclamation_mark: " + alertSeverity
	case "RECOVERY":
		alertSeverity = ":white_check_mark: " + alertSeverity
	}

	if alertSysName == "%SYSNAME%" {
		alertSysName = alertHost
	}

	msg := fmt.Sprintf("### LibreNMS\n\n**%s**\n\n%s",
		alertSeverity,
		alertMsg,
	)

	if alertHost == "%HOST%" {
		DispatchMessage(ctx, msg)
		return
	}

	conf := GetCtxConfig(ctx)
	routeID := GetCtxRouteID(ctx)
	routeConfig := conf.Routes[routeID]

	contactRoutes, exists := routeRegexs[routeID]
	if !exists {
		makeRouteMatches(routeID, routeConfig)

		contactRoutes, exists = routeRegexs[routeID]
		if !exists {
			DispatchMessage(ctx, msg)
			return
		}
	}

	if contactRoutes == nil {
		DispatchMessage(ctx, msg)
		return
	}

	// Custom message routing
	if err := setupLibreNMSClient(routeConfig); err != nil {
		fmt.Println(err.Error())
		return
	}

	dev, err := libreNMSClient.GetDevice(alertHost)
	if err != nil {
		fmt.Println(err)
		return
	}

	if dev == nil {
		fmt.Printf("Couldn't find device '%s', sending to non-routed channels\n", alertHost)
		DispatchMessage(ctx, msg)
		return
	}

	if dev.SysContact == "" {
		fmt.Printf("No sysContact defined for device '%s', sending to non-routed channels\n", alertHost)
		DispatchMessage(ctx, msg)
		return
	}

	b := bot.GetBot()
	for _, c := range contactRoutes {
		if c.match.MatchString(dev.SysContact) {
			if err := b.SendMsgTeamChannel(c.channel, msg); err != nil {
				fmt.Println(err)
			}
		}
	}
}

func makeRouteMatches(id string, rc *config.RouteConfig) {
	contactRoutes, exists := rc.Settings["routes"].(map[string]interface{})
	if !exists {
		routeRegexs[id] = nil
		return
	}

	contacts := make([]contact, 0, len(contactRoutes))
	for email, channel := range contactRoutes {
		if len(email) == 0 {
			continue
		}

		if email == "*" { // Match everything
			email = ".*"
		} else if email[0] != '/' { // No forward slash prefix means literal string
			email = regexp.QuoteMeta(email)
		} else {
			// Must be enclosed in forward slash
			if email[len(email)-1] != '/' {
				fmt.Printf("Invalid regex: %s\n", email)
				continue
			}

			email = email[1 : len(email)-1] // Chop off /.../
		}

		r, err := regexp.Compile(email)
		if err != nil {
			fmt.Printf("Invalid regex: %s\n", email)
			continue
		}

		c := contact{channel: channel.(string), match: r}
		contacts = append(contacts, c)
	}

	routeRegexs[id] = contacts
}

func setupLibreNMSClient(conf *config.RouteConfig) error {
	if libreNMSClient != nil {
		return nil
	}

	address, ok := conf.Settings["address"].(string)
	if !ok {
		return errors.New("bad LibreNMS address")
	}

	c, err := librenms.NewClient(address)
	if err != nil {
		return errors.New("bad LibreNMS address")
	}

	skipVerify, ok := conf.Settings["skip_verify"].(bool)
	if ok && skipVerify {
		c.SkipTLSVerify()
	}

	token, ok := conf.Settings["apitoken"].(string)
	if !ok {
		return errors.New("bad LibreNMS apitoken")
	}

	if err := c.Login(token); err != nil {
		return err
	}

	libreNMSClient = c
	return nil
}
