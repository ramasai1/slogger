package slogger

import (
	"encoding/json"
	"fmt"
	"strconv"
)

func FormatLog(log *Log) string {
	year, month, day := log.Timestamp.Date()
	hour, min, sec := log.Timestamp.Clock()
	millisec := log.Timestamp.Nanosecond() / 1000000

	return formatLog(log, fmt.Sprintf("[%.4d/%.2d/%.2d %.2d:%.2d:%.2d.%.3d]",
		year, month, day,
		hour, min, sec,
		millisec,
	))
}

func formatLog(log *Log, timePart string) string {
	errorCodeStr := ""
	if log.ErrorCode != NoErrorCode {
		errorCodeStr += fmt.Sprintf("[%v] ", log.ErrorCode)
	}

	structuredLog, _ := json.Marshal(map[string]string{
		"time": timePart,
		"prefix": log.Prefix,
		"level": log.Level.Type(),
		"file": log.Filename,
		"func": log.FuncName,
		"line": strconv.Itoa(log.Line),
		"error": strconv.Itoa(int(log.ErrorCode)),
		"msg": log.Message(),
	})

	return string(structuredLog)
}

func convertOffsetToString(offset int) string {
	if offset == 0 {
		return "+0000"
	}
	var sign string
	if offset > 0 {
		sign = "+"
	} else {
		sign = "-"
		offset *= -1
	}
	hoursOffset := float32(offset) / 3600.0
	var leadingZero string
	if hoursOffset > -9 && hoursOffset < 9 {
		leadingZero = "0"
	}
	return fmt.Sprintf("%s%s%.0f", sign, leadingZero, hoursOffset*100.0)
}

func FormatLogWithTimezone(log *Log) string {
	year, month, day := log.Timestamp.Date()
	hour, min, sec := log.Timestamp.Clock()
	millisec := log.Timestamp.Nanosecond() / 1000000
	_, offset := log.Timestamp.Zone() // offset in seconds

	return formatLog(log, fmt.Sprintf("[%.4d-%.2d-%.2dT%.2d:%.2d:%.2d.%.3d%s]",
		year, month, day,
		hour, min, sec,
		millisec,
		convertOffsetToString(offset)),
	)
}
