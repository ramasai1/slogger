package list_appender

import (
	"github.com/mongodb/slogger/v3/slogger"
	"sync"
)

type ListAppender struct {
	slogger.BaseAppender

	lock sync.Mutex
	logs []string
	capacity int
}

func New(capacity int, initialLevel slogger.Level) *ListAppender {
	appender := &ListAppender{
		lock: sync.Mutex{},
		logs: make([]string, capacity),
		capacity: capacity,
	}
	appender.SetLevel(initialLevel)

	return appender
}

func (self *ListAppender) Append(log *slogger.Log) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	if len(self.logs) == self.capacity {
		self.Flush()
	}
	self.logs = append(self.logs, slogger.FormatLog(log))
	return nil
}

func (self *ListAppender) Flush() error {
	self.logs = self.logs[:0]
	return nil
}

func (self *ListAppender) GetLogs() []string {
	return self.logs
}
