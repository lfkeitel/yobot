package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kennygrant/sanitize"

	"github.com/lfkeitel/yobot/ircbot"
)

func main() {}

func init() {
	ircbot.RegisterCommand("#startmeeting", startMeetingCmd)
	ircbot.RegisterCommand("#endmeeting", endMeetingCmd)

	ircbot.RegisterCommand("#addchair", addChairCmd)
	ircbot.RegisterCommand("#rmchair", rmChairCmd)

	ircbot.RegisterCommand("#topic", topicCmd)
	ircbot.RegisterCommand("#agreed", agreedCmd)

	ircbot.RegisterCommand("#info", infoCmd)
	ircbot.RegisterCommand("#action", actionCmd)
	ircbot.RegisterCommand("#link", linkCmd)

	ircbot.RegisterCommand("#rollcall", rollcallCmd)
}

func timeNowInUTC() string {
	return time.Now().In(time.UTC).Format("15:04:05")
}

var startMeetingCmd = &ircbot.Command{
	Help: "Start a meeting: #startmeeting Meeting Name",
	Handler: func(conn *ircbot.Conn, event *ircbot.Event) error {
		if !ircbot.IsChannel(event.Source) {
			return nil
		}

		meetingsLock.Lock()
		defer meetingsLock.Unlock()
		if meet, exists := meetings[event.Source]; exists {
			conn.Privmsgf(event.Source, "This channel already has a meeting: %s", meet.Name)
			return nil
		}

		if len(event.Args) == 0 {
			conn.Privmsg(event.Source, "Help: #startmeeting Meeting Name")
			return nil
		}

		meetingName := strings.Join(event.Args, " ")

		if err := os.MkdirAll(event.Config.ModuleDataDir("meetbot"), 0755); err != nil {
			conn.Privmsg(event.Source, "Error starting meeting. Please see logs.")
			return err
		}

		m := &meeting{
			Started:   time.Now().In(time.UTC),
			Name:      meetingName,
			Channel:   event.Source,
			StartedBy: event.Line.Nick,
			Chairs:    []string{event.Line.Nick},
			Rollcall:  []string{event.Line.Nick},
		}
		meetings[event.Source] = m

		ircbot.RegisterTap(m.tap, "meetbot", event.Source)

		conn.Privmsgf(event.Source, "Meeting started %s. The chair is %s.", m.Started.Format(meetingTimeFormat), event.Line.Nick)
		conn.Privmsg(event.Source, "Useful commands: #action #agreed #info #topic #rollcall")
		conn.Topic(event.Source, fmt.Sprintf("Meeting: %s", meetingName))
		conn.Privmsgf(event.Source, "The meeting name has been set to '%s'", meetingName)

		return nil
	},
}

var endMeetingCmd = &ircbot.Command{
	Help: "End a meeting: #endmeeting",
	Handler: func(conn *ircbot.Conn, event *ircbot.Event) error {
		if !ircbot.IsChannel(event.Source) {
			return nil
		}

		meetingsLock.Lock()
		defer meetingsLock.Unlock()
		meet, exists := meetings[event.Source]
		if !exists {
			conn.Privmsg(event.Source, "This channel doesn't have a meeting.")
			return nil
		}

		if !meet.isChair(event.Line.Nick) {
			conn.Privmsg(event.Source, "Only chairs can end a meeting.")
			return nil
		}

		meet.end(conn, event)
		delete(meetings, event.Source)
		ircbot.UnregisterTap("meetbot", event.Source)

		conn.Topic(event.Source, fmt.Sprintf("Meeting room %s", event.Source))
		conn.Privmsgf(event.Source, "Meeting ended %s.", meet.Ended.Format(meetingTimeFormat))

		meetingPath := filepath.Join(event.Config.ModuleDataDir("meetbot"), sanitize.Name(meet.Name))

		if err := os.MkdirAll(meetingPath, 0755); err != nil {
			conn.Privmsg(event.Source, "Error saving meeting log. Please see application logs.")
			return err
		}

		meetingSummaryPath := filepath.Join(meetingPath, meet.Started.Format(time.RFC3339))
		log, err := os.OpenFile(meetingSummaryPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			conn.Privmsg(event.Source, "Error saving meeting minutes. Please see application logs.")
			return err
		}

		log.Write(meet.buildLog())
		log.Close()

		if err := ioutil.WriteFile(meetingSummaryPath+".log", meet.Log.Bytes(), 0644); err != nil {
			conn.Privmsg(event.Source, "Error saving meeting log. Please see application logs.")
			return err
		}

		return nil
	},
}

var addChairCmd = &ircbot.Command{
	Help: "Add a chair: #addchair nickname",
	Handler: func(conn *ircbot.Conn, event *ircbot.Event) error {
		if !ircbot.IsChannel(event.Source) {
			return nil
		}

		if len(event.Args) != 1 {
			conn.Privmsg(event.Source, "Help: #addchair nickname")
			return nil
		}

		meetingsLock.Lock()
		defer meetingsLock.Unlock()
		meet, exists := meetings[event.Source]
		if !exists {
			conn.Privmsg(event.Source, "This channel doesn't have a meeting.")
			return nil
		}

		if !meet.isChair(event.Line.Nick) {
			conn.Privmsg(event.Source, "Only chairs can add another chair.")
			return nil
		}

		meet.addChair(event.Args[0])

		conn.Privmsgf(event.Source, "Current chairs: %s", strings.Join(meet.Chairs, " "))
		return nil
	},
}

var rmChairCmd = &ircbot.Command{
	Help: "Remove a chair: #rmchair nickname",
	Handler: func(conn *ircbot.Conn, event *ircbot.Event) error {
		if !ircbot.IsChannel(event.Source) {
			return nil
		}

		if len(event.Args) != 1 {
			conn.Privmsg(event.Source, "Help: #rmchair nickname")
			return nil
		}

		meetingsLock.Lock()
		defer meetingsLock.Unlock()
		meet, exists := meetings[event.Source]
		if !exists {
			conn.Privmsg(event.Source, "This channel doesn't have a meeting.")
			return nil
		}

		if !meet.isChair(event.Line.Nick) {
			conn.Privmsg(event.Source, "Only chairs can add another chair.")
			return nil
		}

		if len(meet.Chairs) == 1 {
			conn.Privmsg(event.Source, "At least one chair is required")
			return nil
		}

		meet.rmChair(event.Args[0])

		conn.Privmsgf(event.Source, "Current chairs: %s", strings.Join(meet.Chairs, " "))
		return nil
	},
}

var rollcallCmd = &ircbot.Command{
	Help: "Add yourself to the roll call: #rollcall",
	Handler: func(conn *ircbot.Conn, event *ircbot.Event) error {
		if !ircbot.IsChannel(event.Source) {
			return nil
		}

		meetingsLock.Lock()
		defer meetingsLock.Unlock()
		meet, exists := meetings[event.Source]
		if !exists {
			conn.Privmsg(event.Source, "This channel doesn't have a meeting.")
			return nil
		}

		meet.addRollcall(event.Line.Nick)
		conn.Notice(event.Line.Nick, "You've been added to the rollcall")
		return nil
	},
}

var topicCmd = &ircbot.Command{
	Help: "Set the meeting topic: #topic topic name",
	Handler: func(conn *ircbot.Conn, event *ircbot.Event) error {
		if !ircbot.IsChannel(event.Source) {
			return nil
		}

		if len(event.Args) == 0 {
			conn.Privmsg(event.Source, "Help: #topic topic name")
			return nil
		}

		meetingsLock.Lock()
		defer meetingsLock.Unlock()
		meet, exists := meetings[event.Source]
		if !exists {
			conn.Privmsg(event.Source, "This channel doesn't have a meeting.")
			return nil
		}

		if !meet.isChair(event.Line.Nick) {
			conn.Privmsg(event.Source, "Only chairs can set the meeting topic.")
			return nil
		}

		topicName := strings.Join(event.Args, " ")
		internalName := fmt.Sprintf("%s  (%s, %s)", topicName, event.Nick, timeNowInUTC())

		meet.Topics = append(meet.Topics, topic{Name: internalName})

		conn.Topic(event.Source, fmt.Sprintf("%s (Meeting topic: %s (%s))", topicName, meet.Name, meet.Started.Format("2006-01-02")))
		return nil
	},
}

var actionCmd = &ircbot.Command{
	Help: "Add an action item: #action nick details...",
	Handler: func(conn *ircbot.Conn, event *ircbot.Event) error {
		if !ircbot.IsChannel(event.Source) {
			return nil
		}

		if len(event.Args) < 2 {
			conn.Privmsg(event.Source, "Help: #action nick details...")
			return nil
		}

		meetingsLock.Lock()
		defer meetingsLock.Unlock()
		meet, exists := meetings[event.Source]
		if !exists {
			conn.Privmsg(event.Source, "This channel doesn't have a meeting.")
			return nil
		}

		meet.Actions = append(meet.Actions, action{
			AssignedTo: event.Args[0],
			Action:     strings.Join(event.Args[1:], " "),
		})
		return nil
	},
}

var infoCmd = &ircbot.Command{
	Help:    "Add an info item: #info info details",
	Handler: makeNoteHandler("info"),
}

var agreedCmd = &ircbot.Command{
	Help:    "Add an agreement: #agreed agreement details",
	Handler: makeNoteHandler("agreed"),
}

var linkCmd = &ircbot.Command{
	Help:    "Add an link item: #link link details",
	Handler: makeNoteHandler("link"),
}

func makeNoteHandler(prefix string) ircbot.CommandHandler {
	msgprefix := ""
	if prefix != "info" {
		msgprefix = fmt.Sprintf("%s: ", strings.ToUpper(prefix))
	}

	return func(conn *ircbot.Conn, event *ircbot.Event) error {
		if !ircbot.IsChannel(event.Source) {
			return nil
		}

		if len(event.Args) == 0 {
			conn.Privmsgf(event.Source, "Help: #%s %s details", prefix, prefix)
			return nil
		}

		meetingsLock.Lock()
		defer meetingsLock.Unlock()
		meet, exists := meetings[event.Source]
		if !exists {
			conn.Privmsg(event.Source, "This channel doesn't have a meeting.")
			return nil
		}

		if prefix == "agreed" && !meet.isChair(event.Line.Nick) {
			conn.Privmsg(event.Source, "Only chairs can add an agreement item.")
			return nil
		}

		if len(meet.Topics) == 0 {
			conn.Privmsg(event.Source, "Please set the topic first.")
			return nil
		}

		msg := fmt.Sprintf("%s%s  (%s, %s)", msgprefix, strings.Join(event.Args, " "), event.Nick, timeNowInUTC())
		topicID := len(meet.Topics) - 1
		meet.Topics[topicID].Items = append(meet.Topics[topicID].Items, msg)
		return nil
	}
}
