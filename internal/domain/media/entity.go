package media

import "time"

// Screenshot represents a VRChat screenshot with extracted metadata.
type Screenshot struct {
	ID        string
	FilePath  string
	WorldID   string
	WorldName string
	TakenAt   *time.Time
}
