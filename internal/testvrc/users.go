// Package testvrc provides synthetic VRChat user IDs and display names for tests.
// Do not use real account data in committed tests — see docs/agents/redaction.md.
// IDs use hex UUID segments only (matches production usr_* and media XMP parsers).
package testvrc

const (
	PlayerDisplayName = "TestPlayerAlpha"
	PlayerUserID      = "usr_a1111111-1111-4111-8111-111111111101"
	FriendsHostUserID = "usr_a2222222-2222-4222-8222-222222222202"
	HiddenHostUserID  = "usr_a3333333-3333-4333-8333-333333333303"
	OtherPlayerUserID = "usr_a4444444-4444-4444-8444-444444444404"
	EmbedUserID       = "usr_a5555555-5555-4555-8555-555555555505"
)
