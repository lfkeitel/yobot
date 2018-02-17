package main

import (
	"bytes"
	"fmt"
	"sync"
	"text/template"
	"time"

	irc "github.com/lfkeitel/goirc/client"
	"github.com/lfkeitel/yobot/ircbot"
	"github.com/lfkeitel/yobot/utils"
)

const (
	meetingTimeFormat = "Mon Jan 02 15:04:05 2006 MST"
)

var (
	meetings     = map[string]*meeting{}
	meetingsLock sync.Mutex
)

type meeting struct {
	Started   time.Time
	Ended     time.Time
	Name      string
	Channel   string
	StartedBy string
	Chairs    []string
	Rollcall  []string
	Topics    []topic
	Actions   []action
}

type topic struct {
	Name  string
	Items []string
}

type action struct {
	AssignedTo string
	Action     string
}

func (m *meeting) end(conn *irc.Conn, event *ircbot.Event) {
	m.Ended = time.Now().In(time.UTC)
}

func (m *meeting) addRollcall(nick string) {
	if !utils.StringInSlice(nick, m.Rollcall) {
		m.Rollcall = append(m.Rollcall, nick)
	}
}

func (m *meeting) isChair(nick string) bool {
	return utils.StringInSlice(nick, m.Chairs)
}

func (m *meeting) addChair(nick string) {
	if !utils.StringInSlice(nick, m.Chairs) {
		m.Chairs = append(m.Chairs, nick)
	}
}

func (m *meeting) rmChair(nick string) {
	i := utils.IndexOfString(nick, m.Chairs)
	if i == -1 {
		return
	}
	m.Chairs = append(m.Chairs[:i], m.Chairs[i+1:]...)
}

func (m *meeting) buildLog() []byte {
	var buf bytes.Buffer
	if err := meetingLogTemplate.Execute(&buf, m); err != nil {
		fmt.Println(err)
		return []byte{}
	}
	return buf.Bytes()
}

var meetingLogTemplate = template.Must(template.New("").Parse(`===========================================
{{.Channel}}: {{.Name}}
===========================================

Meeting started by {{.StartedBy}} at {{.Started.Format "15:04:05"}} UTC.


Meeting summary
---------------
{{range .Topics}}* {{.Name}}
{{range .Items}}  * {{.}}
{{end}}{{end}}

Meeting ended at {{.Ended.Format "15:04:05"}} UTC.


Action Items
------------
{{range .Actions}}* {{.AssignedTo}}: {{.Action}}
{{end}}

Rollcall
--------
{{range .Rollcall}}* {{.}}
{{end}}
`))
