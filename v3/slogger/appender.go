package slogger

type Appender interface {
	Append(log *Log) error
	Flush() error
	Allow(level Level) bool
	SetLevel(level Level)
	GetLevel() Level
}
