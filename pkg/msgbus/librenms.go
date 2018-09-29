package msgbus

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/lfkeitel/yobot/librenms"
	"github.com/lfkeitel/yobot/pkg/bot"
	"github.com/lfkeitel/yobot/pkg/config"
	"github.com/lfkeitel/yobot/pkg/utils"
)

func init() {
	RegisterMsgBus("librenms", handleLibreNMS)
}

var libreNMSClient *librenms.Client

func handleLibreNMS(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	alertTitle := utils.StringOrDefault(r.Form.Get("title"), "%TITLE%")
	alertHost := utils.StringOrDefault(r.Form.Get("host"), "%HOST%")
	alertSysName := utils.StringOrDefault(r.Form.Get("sysName"), "%SYSNAME%")
	alertRuleName := utils.StringOrDefault(r.Form.Get("rule"), "%RULE%")
	alertTimestamp := utils.StringOrDefault(r.Form.Get("timestamp"), "%TIMESTAMP%")
	alertSeverity := strings.ToUpper(utils.StringOrDefault(r.Form.Get("severity"), "CRITICAL"))

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

	msg := fmt.Sprintf("### LibreNMS\n\n**%s** - %s on host %s - %s @ %s",
		alertSeverity,
		alertTitle,
		alertHost,
		alertRuleName,
		alertTimestamp,
	)

	conf := GetCtxConfig(ctx)
	routeID := GetCtxRouteID(ctx)

	routeConfig := conf.Routes[routeID]
	contactRoutes, exists := routeConfig.Settings["routes"].(map[string]interface{})
	if !exists {
		DispatchMessage(ctx, msg)
		return
	}

	// Custom message routing
	if err := setupLibreNMSClient(routeConfig); err != nil {
		fmt.Println(err.Error())
		return
	}

	dev, err := libreNMSClient.GetDevice(alertSysName)
	if err != nil {
		fmt.Println(err)
		return
	}

	if dev == nil {
		fmt.Printf("Couldn't find device '%s', sending to non-routed channels\n", alertSysName)
		DispatchMessage(ctx, msg)
		return
	}

	b := bot.GetBot()
	for email, channel := range contactRoutes {
		if email == "*" || strings.Contains(dev.SysContact, email) {
			if err := b.SendMsgTeamChannel(channel.(string), msg); err != nil {
				fmt.Println(err)
			}
		}
	}
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
