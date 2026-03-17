package settings

import "time"

// AppSetting represents a key-value app setting.
type AppSetting struct {
	Key       string
	Value     string
	UpdatedAt *time.Time
}
