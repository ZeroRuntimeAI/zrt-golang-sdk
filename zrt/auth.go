package zrt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// ResolveAuthToken resolves a ZRT auth token from an explicit value, then
// ZRT_AUTH_TOKEN, then minting from ZRT_API_KEY + ZRT_SECRET_KEY.
func ResolveAuthToken(explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	if pre := os.Getenv("ZRT_AUTH_TOKEN"); pre != "" {
		return pre, nil
	}
	apiKey := os.Getenv("ZRT_API_KEY")
	secret := os.Getenv("ZRT_SECRET_KEY")
	if apiKey != "" && secret != "" {
		tok, err := mintJWT(apiKey, secret)
		if err != nil {
			return "", fmt.Errorf("%w: minting JWT from ZRT_API_KEY+ZRT_SECRET_KEY: %w. Fix the keys or set ZRT_AUTH_TOKEN directly", ErrAuthFailed, err)
		}
		return tok, nil
	}
	return "", fmt.Errorf("%w: set ZRT_AUTH_TOKEN, or ZRT_API_KEY + ZRT_SECRET_KEY, or pass an auth token in RoomOptions", ErrNoCredentials)
}

func mintJWT(apiKey, secret string) (string, error) {
	// Structs (not maps) so JSON field order on the wire is deterministic.
	header := struct {
		Alg string `json:"alg"`
		Typ string `json:"typ"`
	}{Alg: "HS256", Typ: "JWT"}
	now := time.Now().Unix()
	payload := struct {
		APIKey      string   `json:"apikey"`
		Permissions []string `json:"permissions"`
		IAT         int64    `json:"iat"`
		EXP         int64    `json:"exp"`
	}{APIKey: apiKey, Permissions: []string{"allow_join"}, IAT: now, EXP: now + 24*60*60}
	hb, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	pb, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	headerB64 := b64url(hb)
	payloadB64 := b64url(pb)
	signingInput := headerB64 + "." + payloadB64
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingInput))
	sig := mac.Sum(nil)
	return signingInput + "." + b64url(sig), nil
}

func b64url(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}
