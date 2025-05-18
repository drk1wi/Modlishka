/**

    "Modlishka" Reverse Proxy.

    Copyright 2018 (C) Piotr Duszyński piotr[at]duszynski.eu. All rights reserved.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.

    You should have received a copy of the Modlishka License along with this program.

**/

package log

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type Handler func(format string, args ...interface{})

// https://misc.flogisoft.com/bash/tip_colors_and_formatting
const (
	BOLD = "\033[1m"
	DIM  = "\033[2m"

	FG_BLACK = "\033[30m"
	FG_WHITE = "\033[97m"

	BG_DGRAY  = "\033[100m"
	BG_RED    = "\033[41m"
	BG_GREEN  = "\033[42m"
	BG_YELLOW = "\033[43m"

	RESET = "\033[0m"
)

const (
	DEBUG = iota
	STAT
	INFO
	WARNING
	ERROR
	FATAL
)

var (
	WithColors = true
	Output     = os.Stderr
	DateFormat = "Mon Jan _2 15:04:05 2006"
	MinLevel   = DEBUG

	mutex  = &sync.Mutex{}
	labels = map[int]string{
		DEBUG:   "DBG",
		INFO:    "INF",
		WARNING: "WAR",
		ERROR:   "ERR",
		STAT:    "STAT",
		FATAL:   "!!!",
	}
	colors = map[int]string{
		DEBUG:   DIM + FG_BLACK + BG_DGRAY,
		STAT:    DIM + FG_BLACK + BG_DGRAY,
		INFO:    FG_WHITE + BG_GREEN,
		WARNING: FG_WHITE + BG_YELLOW,
		ERROR:   FG_WHITE + BG_RED,
		FATAL:   FG_WHITE + BG_RED + BOLD,
	}

	Options LoggingOptions
)

type LoggingOptions struct {
	GET            bool
	POST           bool
	LogRequestPath string
}

func Wrap(s, effect string) string {
	if WithColors == true {
		s = effect + s + RESET
	}
	return s
}

func Dim(s string) string {
	return Wrap(s, DIM)
}

func Log(level int, format string, args ...interface{}) {
	if level >= MinLevel {

		label := labels[level]
		color := colors[level]
		when := time.Now().UTC().Format(DateFormat)

		what := fmt.Sprintf(format, args...)
		if strings.HasSuffix(what, "\n") == false {
			what += "\n"
		}

		l := Dim("[%s]")
		r := Wrap(" %s ", color) + " %s"

		fmt.Fprintf(Output, l+" "+r, when, label, what)
	}
}

func Debugf(format string, args ...interface{}) {
	Log(DEBUG, format, args...)

}

func Infof(format string, args ...interface{}) {
	Log(INFO, format, args...)
}

func Warningf(format string, args ...interface{}) {
	Log(WARNING, format, args...)
}

func Errorf(format string, args ...interface{}) {
	Log(ERROR, format, args...)
}

func Statf(format string, args ...interface{}) {
	Log(INFO, format, args...)
}

func Fatalf(format string, args ...interface{}) {
	Log(FATAL, format, args...)
	os.Exit(1)
}

func Fatal(err error) {
	Log(FATAL, "%s", err)
	os.Exit(1)
}
