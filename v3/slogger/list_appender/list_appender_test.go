package list_appender

import (
	"github.com/mongodb/slogger/v3/slogger"
	"strings"
	"testing"
)

func TestLog(test *testing.T) {
	logger, appender := setup(slogger.INFO)

	logMsg := "this is a test log message"
	logger.Logf(slogger.WARN, logMsg)
	logFound := false
	for _, log := range appender.GetLogs() {
		if strings.Contains(log, logMsg) {
			logFound = true
			break
		}
	}

	if !logFound {
		test.Errorf("Log did not make it through even though appender was set to INFO")
	}

	// Change level and attempt logging
	appender.Flush()
	appender.SetLevel(slogger.WARN)
	logger.Logf(slogger.INFO, logMsg)
	logFound = false
	for _, log := range appender.GetLogs() {
		if strings.Contains(log, logMsg) {
			logFound = true
			break
		}
	}

	if logFound {
		test.Errorf("Log made it through -- should NOT have made it through because appender level is WARN")
	}
}

func setup(initialLogLevel slogger.Level) (slogger.Logger, *ListAppender){
	listAppender := New(10, initialLogLevel)
	logger := slogger.Logger{
		Appenders: []slogger.Appender{listAppender},
	}

	return logger, listAppender
}
