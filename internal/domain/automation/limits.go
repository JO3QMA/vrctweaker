package automation

import "time"

const (
	RunLogMaxEntries       = 50
	MaxActionsPerItem      = 10
	MaxScriptBytes         = 32 * 1024
	LuaExecTimeout         = 10 * time.Second
	ScheduleTickResolution = time.Minute
	EventQueueCapacity     = 256
	FailureLogRateLimit    = 10 * time.Minute
)
