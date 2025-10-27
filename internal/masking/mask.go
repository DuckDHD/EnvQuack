package masking

import (
	"regexp"
	"strings"
)

// disableMasking allows advanced users to disable masking for debugging
var disableMasking bool

// Sensitive patterns to check in variable names
var sensitivePatterns = []string{
	"PASSWORD",
	"PASSWD",
	"SECRET",
	"KEY",
	"TOKEN",
	"API_KEY",
	"APIKEY",
	"PRIVATE",
	"CREDENTIAL",
	"CRED",
	"AUTH",
	"CERTIFICATE",
	"CERT",
	"OAUTH",
	"JWT",
	"SESSION",
	"SALT",
	"HASH",
	"SIGNATURE",
	"ENCRYPTION",
	"DECRYPT",
}

// Common secret prefixes that indicate sensitive data
var sensitivePrefixes = []string{
	"sk_",     // Stripe secret keys
	"pk_",     // Stripe public keys
	"rk_",     // Stripe restricted keys
	"ghp_",    // GitHub personal access token
	"gho_",    // GitHub OAuth token
	"ghs_",    // GitHub server-to-server token
	"ghr_",    // GitHub refresh token
	"AKIA",    // AWS access key
	"ASIA",    // AWS session token
	"xoxb-",   // Slack bot token
	"xoxp-",   // Slack user token
	"xoxa-",   // Slack app token
	"AIza",    // Google API key
	"ya29.",   // Google OAuth token
	"sqo_",    // Square OAuth
	"key-",    // Generic key prefix
	"token-",  // Generic token prefix
	"Bearer ", // Bearer tokens
}

// Regex patterns for sensitive value formats
var sensitiveValuePatterns = []*regexp.Regexp{
	// JWT tokens (3 parts separated by dots)
	regexp.MustCompile(`^[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+$`),

	// Base64-encoded secrets (long alphanumeric strings)
	regexp.MustCompile(`^[A-Za-z0-9+/]{32,}={0,2}$`),

	// Hex-encoded secrets (long hex strings)
	regexp.MustCompile(`^[a-fA-F0-9]{32,}$`),

	// UUIDs (might be used as secrets)
	regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`),
}

// SetMaskingEnabled controls whether masking is enabled
func SetMaskingEnabled(enabled bool) {
	disableMasking = !enabled
}

// IsSensitive checks if a variable name indicates sensitive content
func IsSensitive(key string) bool {
	upperKey := strings.ToUpper(key)

	// Check if key contains any sensitive pattern
	for _, pattern := range sensitivePatterns {
		if strings.Contains(upperKey, pattern) {
			return true
		}
	}

	return false
}

// hasSensitivePrefix checks if a value has a known sensitive prefix
func hasSensitivePrefix(value string) bool {
	for _, prefix := range sensitivePrefixes {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}

// looksLikeSecret checks if a value matches patterns of common secret formats
func looksLikeSecret(value string) bool {
	// Skip short values
	if len(value) < 16 {
		return false
	}

	// Check against patterns
	for _, pattern := range sensitiveValuePatterns {
		if pattern.MatchString(value) {
			return true
		}
	}

	return false
}

// Mask replaces sensitive parts of a value with asterisks
// Preserves first 2 and last 2 characters for recognition
func Mask(key, value string) string {
	// Handle empty values
	if value == "" {
		return "(empty)"
	}

	// Handle very short values (4 chars or less)
	if len(value) <= 4 {
		return "***"
	}

	// For longer values, show first 2 and last 2 chars
	if len(value) <= 8 {
		// Short but not too short: show first and last char
		return string(value[0]) + "***" + string(value[len(value)-1])
	}

	// Standard masking: first 2, asterisks, last 2
	first := value[:2]
	last := value[len(value)-2:]

	// Calculate middle length (but cap at reasonable display)
	middleLen := len(value) - 4
	if middleLen > 20 {
		middleLen = 20 // Cap for readability
	}

	middle := strings.Repeat("*", middleLen)

	return first + middle + last
}

// MaskIfSensitive masks value only if key or value indicates it's sensitive
func MaskIfSensitive(key, value string) string {
	// If masking is disabled, return value as-is
	if disableMasking {
		return value
	}

	// Check key name
	if IsSensitive(key) {
		return Mask(key, value)
	}

	// Check value prefix (GitHub tokens, Stripe keys, etc.)
	if hasSensitivePrefix(value) {
		return Mask(key, value)
	}

	// Check if value looks like a secret
	if looksLikeSecret(value) {
		return Mask(key, value)
	}

	return value
}

// MaskMap masks all sensitive values in a map
func MaskMap(vars map[string]string) map[string]string {
	masked := make(map[string]string, len(vars))
	for key, value := range vars {
		masked[key] = MaskIfSensitive(key, value)
	}
	return masked
}
