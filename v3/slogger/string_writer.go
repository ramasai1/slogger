package slogger

type StringWriter interface {
	WriteString(s string) (ret int, err error)
	Sync() error
}
