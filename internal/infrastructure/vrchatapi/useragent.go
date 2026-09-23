package vrchatapi

import "vrchat-tweaker/internal/appversion"

// UserAgent is sent on every VRChat API and Pipeline WebSocket request. VRChat requires a
// properly formatted User-Agent with application name, version, and contact information.
// See https://vrchat.community/reference/get-current-user
func UserAgent() string {
	return appversion.UserAgent()
}
