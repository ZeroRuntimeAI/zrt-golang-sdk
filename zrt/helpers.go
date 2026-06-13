package zrt

import "os"

// Ptr returns a pointer to v. Handy for optional plugin fields.
func Ptr[T any](v T) *T { return &v }

// String returns a pointer to v, for optional string fields.
// It reads more clearly than Ptr at call sites (zrt.String("x")).
func String(v string) *string { return &v }

// Bool returns a pointer to v, for optional bool fields.
func Bool(v bool) *bool { return &v }

// Int returns a pointer to v, for optional int fields.
func Int(v int) *int { return &v }

// Float64 returns a pointer to v, for optional float64 fields
// (zrt.Float64(0.4) reads more clearly than zrt.Ptr(0.4)).
func Float64(v float64) *float64 { return &v }

// StrOr returns s if non-empty, else def.
func StrOr(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

// IntOr dereferences p, or returns def if p is nil.
func IntOr(p *int, def int) int {
	if p == nil {
		return def
	}
	return *p
}

// BoolOr dereferences p, or returns def if p is nil.
func BoolOr(p *bool, def bool) bool {
	if p == nil {
		return def
	}
	return *p
}

// FloatOr dereferences p, or returns def if p is nil.
func FloatOr(p *float64, def float64) float64 {
	if p == nil {
		return def
	}
	return *p
}

// EnvOr returns the value of env var key, or fallback if unset/empty.
func EnvOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// FloatStr formats a float so the result always contains a decimal point
// (for example "1.0" rather than "1"), for plugins building string param maps.
func FloatStr(f float64) string { return floatStr(f) }

// APIKeyOr returns explicit if non-empty, else the value of env var envKey.
func APIKeyOr(explicit, envKey string) string {
	if explicit != "" {
		return explicit
	}
	return os.Getenv(envKey)
}
