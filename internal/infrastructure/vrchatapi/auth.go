package vrchatapi

import "context"

// VRChatAPIClient defines the VRChat API operations used by IdentityUseCase (for testability).
type VRChatAPIClient interface {
	Login(ctx context.Context, username, password, twoFactorCode string) (string, error)
	SetAuthToken(token string)
	// GetAuthToken returns the current in-memory auth token. Empty string means no active session.
	GetAuthToken() string
	GetCurrentUser(ctx context.Context) (*CurrentUserProfile, error)
	GetFriends(ctx context.Context) ([]Friend, error)
	GetUser(ctx context.Context, userID string) (*Friend, error)
	SetUserStatus(ctx context.Context, userID string, status UserStatus) error
	SetUserStatusDescription(ctx context.Context, userID string, description string) error
	SetUserStatusAndDescription(ctx context.Context, userID string, status UserStatus, description string) error
}

// CredentialStore defines storage for auth tokens (OS keyring integration).
type CredentialStore interface {
	// Get retrieves the auth token for the service.
	Get(service, user string) (string, error)
	// Set stores the auth token.
	Set(service, user, password string) error
	// Delete removes the auth token.
	Delete(service, user string) error
}

const CredentialService = "vrchat-tweaker"
const CredentialUser = "auth-token"
