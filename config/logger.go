package config

import (
	"log"
	"time"
)

var (
	blue   = string([]byte{27, 91, 57, 55, 59, 52, 52, 109}) // INF
	yellow = string([]byte{27, 91, 57, 55, 59, 52, 51, 109}) // WRN
	red    = string([]byte{27, 91, 57, 55, 59, 52, 49, 109}) // ERR
	cyan   = string([]byte{27, 91, 57, 55, 59, 52, 54, 109}) // ???
	reset  = string([]byte{27, 91, 48, 109})
)

var disableColor = false

func LogV(logMessage string) {
	Log("VRB", logMessage)
}

func LogW(logMessage string) {
	Log("WRN", logMessage)
}

func LogE(logMessage string) {
	Log("ERR", logMessage)
}

func LogI(logMessage string) {
	Log("INF", logMessage)
}

func Log(logType string, logMessage string) {

	var color string
	switch logType {
	case "VRB":
		color = reset
		break
	case "INF":
		color = blue
		break
	case "WRN":
		color = yellow
		break
	case "ERR":
		color = red
		break
	default:
		color = cyan
	}

	log.Printf("[Kumuluz-config] %v |%s %1s %s| %s\n",
		time.Now().Format("2006/01/02 15:04:05"),
		color, logType, reset,
		logMessage,
	)
}
