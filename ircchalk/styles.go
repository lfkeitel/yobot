package ircchalk

import "strings"

// A Color is an available IRC color
type Color string

// Available IRC colors
const (
	color     = "\x03"
	bold      = "\x02"
	italics   = "\x1D"
	underline = "\x1F"
	reverse   = "\x16"
	reset     = "\x0F"

	White   Color = "00"
	Black   Color = "01"
	Blue    Color = "02"
	Green   Color = "03"
	Red     Color = "04"
	Brown   Color = "05"
	Purple  Color = "06"
	Orange  Color = "07"
	Yellow  Color = "08"
	Lime    Color = "09"
	Teal    Color = "10"
	Cyan    Color = "11"
	Royal   Color = "12"
	Pink    Color = "13"
	Grey    Color = "14"
	Silver  Color = "15"
	Default Color = "99"
)

// Bold text, s is concatinated with spaces
func Bold(s ...string) string {
	return bold + strings.Join(s, " ") + bold
}

// Italics italisizes text, s is concatinated with spaces
func Italics(s ...string) string {
	return italics + strings.Join(s, " ") + italics
}

// Underline text, s is concatinated with spaces
func Underline(s ...string) string {
	return underline + strings.Join(s, " ") + underline
}

// Reset all formatting, s is concatinated with spaces
func Reset(s ...string) string {
	return reset + strings.Join(s, " ") + reset
}

// ReverseColor reverses foreground and background colors, s is concatinated with spaces
func ReverseColor(s ...string) string {
	return reverse + strings.Join(s, " ") + reverse
}

// Chalk colors text with a foreground and background color. s is concatinated with spaces.
// If fore or back are empty strings, they will be the client default
func Chalk(fore, back Color, s ...string) string {
	if fore == "" {
		fore = Default
	}
	if back == "" {
		back = Default
	}

	c := color + string(fore) + "," + string(back)
	return c + strings.Join(s, " ") + string(color)
}

// ChalkFore colors text with a foreground color. s is concatinated with spaces.
// If c is an empty string, the client's default color is used.
func ChalkFore(c Color, s ...string) string {
	if c == "" {
		c = Default
	}

	return color + string(c) + strings.Join(s, " ") + string(color)
}

// ChalkBack colors text with a background color. s is concatinated with spaces.
// If c is an empty string, the client's default color is used.
func ChalkBack(c Color, s ...string) string {
	if c == "" {
		c = Default
	}

	return color + "," + string(c) + strings.Join(s, " ") + string(color)
}
