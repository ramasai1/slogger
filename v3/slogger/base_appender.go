package slogger

import "sync"

type BaseAppender struct {
	minAllowedLogLevel Level
	minAllowedLogLevelLock sync.RWMutex
}

func (self *BaseAppender) Allow(level Level) bool {
	return level >= self.minAllowedLogLevel
}

func (self *BaseAppender) SetLevel(level Level) {
	self.minAllowedLogLevelLock.Lock()
	defer self.minAllowedLogLevelLock.Unlock()

	self.minAllowedLogLevel = level
}

func (self *BaseAppender) GetLevel() Level {
	self.minAllowedLogLevelLock.RLock()
	defer self.minAllowedLogLevelLock.RUnlock()

	return self.minAllowedLogLevel
}
