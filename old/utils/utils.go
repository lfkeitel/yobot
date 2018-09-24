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

func StringInSlice(s string, ss []string) bool {
	for _, v := range ss {
		if s == v {
			return true
		}
	}
	return false
}

func IndexOfString(s string, ss []string) int {
	for i, v := range ss {
		if s == v {
			return i
		}
	}
	return -1
}
