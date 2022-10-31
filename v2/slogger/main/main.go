package main

import "github.com/mongodb/slogger/v2/slogger"

func main() {
	logger := &slogger.Logger{
		Appenders: []slogger.Appender{slogger.StdOutAppender()},
		LevelsToSkip: 2,
	}
	dummyOne(logger, "this is a test log with a request to skip two levels")
}

func dummyOne(logger *slogger.Logger, msg string) {
	dummyTwo(logger, msg)
}

func dummyTwo(logger *slogger.Logger, msg string) {
	err := logger.Errorf(slogger.INFO, msg)
	if err != nil {
		return 
	}
}
