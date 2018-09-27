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
)

func init() {
	RegisterMsgBus("librenms", handleLibreNMS)
}

var libreNMSClient *librenms.Client

func handleLibreNMS(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	severity := strings.ToUpper(r.Form.Get("severity"))

	if strings.Contains(r.Form.Get("title"), "recovered") {
		severity = "RECOVERY"
	}

	switch severity {
	case "CRITICAL":
		severity = ":bangbang: " + severity
	case "WARNING":
		severity = ":heavy_exclamation_mark: " + severity
	case "RECOVERY":
		severity = ":white_check_mark: " + severity
	}

	msg := fmt.Sprintf("### LibreNMS\n\n**%s** - %s on host %s - %s @ %s",
		severity,
		r.Form.Get("title"),
		r.Form.Get("host"),
		r.Form.Get("rule"),
		r.Form.Get("timestamp"),
	)

	conf := GetCtxConfig(ctx)
	routeConfig := conf.Routes["librenms"]
	contactRoutes, exists := routeConfig.Settings["routes"].(map[string]interface{})
	if !exists {
		DispatchMessage(ctx, msg)
		w.Write([]byte(`{"accepted": true}`))
		return
	}

	// Custom message routing
	if err := setupLibreNMSClient(routeConfig); err != nil {
		fmt.Println(err.Error())
		return
	}

	dev, err := libreNMSClient.GetDevice(r.Form.Get("host"))
	if err != nil {
		fmt.Println(err)
		return
	}

	if dev == nil {
		fmt.Println("No device found, sending to non-routed channels")
		DispatchMessage(ctx, msg)
		w.Write([]byte(`{"accepted": true}`))
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
	w.Write([]byte(`{"accepted": true}`))
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
