package usecase

// RejoinTarget is the server-side Instance rejoin destination (not exposed to the frontend).
type RejoinTarget struct {
	PlaySessionID    string
	InstanceID       string
	WorldDisplayName string
}
