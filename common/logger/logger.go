package logger

import (
	log "github.com/sirupsen/logrus"
)

const TimestampFormat = "2006-01-02 15:04:05.999"

func init() {
	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat: TimestampFormat,
	})

	log.SetLevel(log.DebugLevel)
}
