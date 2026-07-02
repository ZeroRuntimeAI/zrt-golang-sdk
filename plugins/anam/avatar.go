// Package anam provides the Anam video-avatar provider.
package anam

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// defaultAvatarID is used when neither AvatarID nor ANAM_AVATAR_ID is set.
const defaultAvatarID = "960f614f-ea88-47c3-9883-f02094f70874"

// Avatar is the Anam video-avatar descriptor.
type Avatar struct {
	zrt.BaseAvatar
	// AvatarID is the Anam avatar identifier.
	AvatarID string
	// PersonaName is the Anam persona name.
	PersonaName string
	// VoiceID is the Anam voice identifier.
	VoiceID string
}

// AvatarOptions configures an Anam Avatar.
type AvatarOptions struct {
	// APIKey overrides the ANAM_API_KEY environment variable.
	APIKey string
	// AvatarID falls back to ANAM_AVATAR_ID, then the default avatar.
	AvatarID string
	// PersonaName is the Anam persona name.
	PersonaName string
	// VoiceID is the Anam voice identifier.
	VoiceID string
}

// NewAvatar builds an Anam Avatar.
func NewAvatar(opts AvatarOptions) *Avatar {
	a := &Avatar{
		AvatarID:    zrt.StrOr(zrt.APIKeyOr(opts.AvatarID, "ANAM_AVATAR_ID"), defaultAvatarID),
		PersonaName: opts.PersonaName,
		VoiceID:     opts.VoiceID,
	}
	a.Init("anam", zrt.APIKeyOr(opts.APIKey, "ANAM_API_KEY"))
	return a
}

// AvatarConfig implements zrt.Avatar.
func (a *Avatar) AvatarConfig() zrt.AvatarRuntimeConfig {
	return zrt.AvatarRuntimeConfig{Provider: "anam", AvatarID: a.AvatarID}
}
