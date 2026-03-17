package launcher

import "time"

// LaunchProfile represents a VRChat launch configuration profile.
type LaunchProfile struct {
	ID        string
	Name      string
	Arguments string
	IsDefault bool
	CreatedAt *time.Time
	UpdatedAt *time.Time
}
