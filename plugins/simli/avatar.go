// Package simli provides the Simli video-avatar provider.
package simli

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// defaultFaceID is used when neither FaceID nor SIMLI_FACE_ID is set.
const defaultFaceID = "afdb6a3e-3939-40aa-92df-01604c23101c"

// Avatar is the Simli video-avatar descriptor.
type Avatar struct {
	zrt.BaseAvatar
	FaceID    string
	Model     string
	IsTrinity bool
}

// AvatarOptions configures a Simli Avatar.
type AvatarOptions struct {
	// APIKey overrides the SIMLI_API_KEY environment variable.
	APIKey string
	// FaceID falls back to SIMLI_FACE_ID, then the default face.
	FaceID    string
	Model     string
	IsTrinity bool
}

// NewAvatar builds a Simli Avatar.
func NewAvatar(opts AvatarOptions) *Avatar {
	a := &Avatar{
		FaceID:    zrt.StrOr(zrt.APIKeyOr(opts.FaceID, "SIMLI_FACE_ID"), defaultFaceID),
		Model:     opts.Model,
		IsTrinity: opts.IsTrinity,
	}
	a.Init("simli", zrt.APIKeyOr(opts.APIKey, "SIMLI_API_KEY"))
	return a
}

// AvatarConfig implements zrt.Avatar.
func (a *Avatar) AvatarConfig() zrt.AvatarRuntimeConfig {
	return zrt.AvatarRuntimeConfig{
		Provider:  "simli",
		FaceID:    a.FaceID,
		Model:     a.Model,
		IsTrinity: a.IsTrinity,
	}
}
