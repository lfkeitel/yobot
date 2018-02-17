package utils

import "strings"

func FirstString(o ...string) string {
	for _, s := range o {
		if s != "" {
			return s
		}
	}

	return ""
}

func FirstLine(s string) string {
	return strings.Split(s, "\n")[0]
}

var quitChan = make(chan bool)

func GetQuitChan() chan bool {
	return quitChan
}
