package logger

import (
	log "github.com/sirupsen/logrus"
	"msa/common/conf"
	"os"
)

const TimestampFormat = "2006-01-02 15:04:05.999"

func init() {
	if conf.Data.Active == conf.ActiveDev {
		log.SetFormatter(&log.TextFormatter{
			TimestampFormat: TimestampFormat,
		})
	} else {
		log.SetFormatter(&log.JSONFormatter{
			TimestampFormat: TimestampFormat,
		})
	}

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)

	// Add the calling method as a field
	// log.SetReportCaller(true)
}
