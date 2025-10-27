package masking

import (
	"strings"
	"testing"
)

func TestIsSensitive(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		sensitive bool
	}{
		// Obvious sensitive keys
		{name: "PASSWORD", key: "PASSWORD", sensitive: true},
		{name: "API_KEY", key: "API_KEY", sensitive: true},
		{name: "SECRET_TOKEN", key: "SECRET_TOKEN", sensitive: true},
		{name: "DATABASE_PASSWORD", key: "DATABASE_PASSWORD", sensitive: true},
		{name: "JWT_SECRET", key: "JWT_SECRET", sensitive: true},
		{name: "STRIPE_SECRET_KEY", key: "STRIPE_SECRET_KEY", sensitive: true},
		{name: "OAUTH_CLIENT_SECRET", key: "OAUTH_CLIENT_SECRET", sensitive: true},
		{name: "ENCRYPTION_KEY", key: "ENCRYPTION_KEY", sensitive: true},
		{name: "PRIVATE_KEY", key: "PRIVATE_KEY", sensitive: true},
		{name: "AUTH_TOKEN", key: "AUTH_TOKEN", sensitive: true},
		{name: "SESSION_SECRET", key: "SESSION_SECRET", sensitive: true},
		{name: "CERTIFICATE", key: "SSL_CERTIFICATE", sensitive: true},
		{name: "CREDENTIAL", key: "DB_CREDENTIAL", sensitive: true},
		{name: "SALT", key: "PASSWORD_SALT", sensitive: true},
		{name: "HASH", key: "SECRET_HASH", sensitive: true},

		// Case variations
		{name: "lowercase password", key: "password", sensitive: true},
		{name: "MixedCase ApiKey", key: "ApiKey", sensitive: true},
		{name: "PASSWD", key: "PASSWD", sensitive: true},

		// Partial matches
		{name: "contains PASSWORD", key: "USER_PASSWORD_HASH", sensitive: true},
		{name: "contains SECRET", key: "MY_SECRET_VALUE", sensitive: true},
		{name: "contains TOKEN", key: "ACCESS_TOKEN_EXPIRY", sensitive: true},

		// Non-sensitive keys
		{name: "DATABASE_URL", key: "DATABASE_URL", sensitive: false},
		{name: "PORT", key: "PORT", sensitive: false},
		{name: "DEBUG", key: "DEBUG", sensitive: false},
		{name: "NODE_ENV", key: "NODE_ENV", sensitive: false},
		{name: "LOG_LEVEL", key: "LOG_LEVEL", sensitive: false},
		{name: "APP_NAME", key: "APP_NAME", sensitive: false},
		{name: "HOST", key: "HOST", sensitive: false},
		{name: "TIMEOUT", key: "TIMEOUT", sensitive: false},
		{name: "MAX_CONNECTIONS", key: "MAX_CONNECTIONS", sensitive: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsSensitive(tt.key)
			if got != tt.sensitive {
				t.Errorf("IsSensitive(%q) = %v, want %v", tt.key, got, tt.sensitive)
			}
		})
	}
}

func TestHasSensitivePrefix(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		sensitive bool
	}{
		// Stripe keys
		{name: "Stripe secret key", value: "sk_live_a1b2c3d4e5f6", sensitive: true},
		{name: "Stripe test key", value: "sk_test_xyz123", sensitive: true},
		{name: "Stripe public key", value: "pk_live_abc123", sensitive: true},
		{name: "Stripe restricted key", value: "rk_live_xyz789", sensitive: true},

		// GitHub tokens
		{name: "GitHub PAT", value: "ghp_a1b2c3d4e5f6g7h8i9j0", sensitive: true},
		{name: "GitHub OAuth", value: "gho_xyz123abc456", sensitive: true},
		{name: "GitHub server token", value: "ghs_abc123def456", sensitive: true},
		{name: "GitHub refresh token", value: "ghr_xyz789abc123", sensitive: true},

		// AWS keys
		{name: "AWS access key", value: "AKIAIOSFODNN7EXAMPLE", sensitive: true},
		{name: "AWS session token", value: "ASIAIOSFODNN7EXAMPLE", sensitive: true},

		// Slack tokens
		{name: "Slack bot token", value: "xoxb-123456789-abcdef", sensitive: true},
		{name: "Slack user token", value: "xoxp-123456789-abcdef", sensitive: true},
		{name: "Slack app token", value: "xoxa-123456789-abcdef", sensitive: true},

		// Google
		{name: "Google API key", value: "AIzaSyA1B2C3D4E5F6G7", sensitive: true},
		{name: "Google OAuth", value: "ya29.a0AfH6SMBx...", sensitive: true},

		// Square
		{name: "Square OAuth", value: "sqo_abc123xyz789", sensitive: true},

		// Generic
		{name: "Generic key prefix", value: "key-abc123xyz789", sensitive: true},
		{name: "Generic token prefix", value: "token-abc123xyz789", sensitive: true},
		{name: "Bearer token", value: "Bearer abc123xyz789", sensitive: true},

		// Non-sensitive
		{name: "URL", value: "https://example.com", sensitive: false},
		{name: "Regular text", value: "some_value", sensitive: false},
		{name: "Number", value: "12345", sensitive: false},
		{name: "Path", value: "/usr/local/bin", sensitive: false},
		{name: "Email", value: "user@example.com", sensitive: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasSensitivePrefix(tt.value)
			if got != tt.sensitive {
				t.Errorf("hasSensitivePrefix(%q) = %v, want %v", tt.value, got, tt.sensitive)
			}
		})
	}
}

func TestLooksLikeSecret(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		isSecret bool
	}{
		// JWT tokens
		{
			name:     "JWT token",
			value:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
			isSecret: true,
		},
		{
			name:     "Short JWT-like (not secret)",
			value:    "abc.def.ghi",
			isSecret: false, // Too short
		},

		// Base64 secrets
		{
			name:     "Base64 secret",
			value:    "YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXoxMjM0NTY3ODkw",
			isSecret: true,
		},
		{
			name:     "Base64 with padding",
			value:    "YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXoxMjM0NTY3ODkwMTIzNDU2Nzg5MA==",
			isSecret: true,
		},

		// Hex secrets
		{
			name:     "Hex secret",
			value:    "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4",
			isSecret: true,
		},
		{
			name:     "Uppercase hex secret",
			value:    "A1B2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C1D2E3F4",
			isSecret: true,
		},

		// UUID (might be secret)
		{
			name:     "UUID",
			value:    "550e8400-e29b-41d4-a716-446655440000",
			isSecret: true,
		},
		{
			name:     "UUID uppercase",
			value:    "550E8400-E29B-41D4-A716-446655440000",
			isSecret: true,
		},

		// Not secrets
		{
			name:     "short value",
			value:    "short",
			isSecret: false,
		},
		{
			name:     "URL",
			value:    "https://example.com/api/v1/endpoint",
			isSecret: false,
		},
		{
			name:     "regular text",
			value:    "some_configuration_value",
			isSecret: false,
		},
		{
			name:     "number string",
			value:    "1234567890",
			isSecret: false,
		},
		{
			name:     "path",
			value:    "/usr/local/bin/application",
			isSecret: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := looksLikeSecret(tt.value)
			if got != tt.isSecret {
				t.Errorf("looksLikeSecret(%q) = %v, want %v", tt.value, got, tt.isSecret)
			}
		})
	}
}

func TestMask(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		expected string
	}{
		{
			name:     "empty value",
			key:      "PASSWORD",
			value:    "",
			expected: "(empty)",
		},
		{
			name:     "very short value (1 char)",
			key:      "KEY",
			value:    "a",
			expected: "***",
		},
		{
			name:     "very short value (4 chars)",
			key:      "KEY",
			value:    "1234",
			expected: "***",
		},
		{
			name:     "short value (5 chars)",
			key:      "TOKEN",
			value:    "abcde",
			expected: "a***e",
		},
		{
			name:     "short value (8 chars)",
			key:      "TOKEN",
			value:    "abc12345",
			expected: "a***5",
		},
		{
			name:     "medium value (14 chars)",
			key:      "API_KEY",
			value:    "sk_live_abc123",
			expected: "sk**********23",
		},
		{
			name:     "long value (32 chars)",
			key:      "SECRET",
			value:    "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
			expected: "a1********************p6",
		},
		{
			name:     "very long value (60 chars) - capped asterisks",
			key:      "LONG_SECRET",
			value:    "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWX",
			expected: "ab********************WX",
		},
		{
			name:     "GitHub token",
			key:      "GITHUB_TOKEN",
			value:    "ghp_a1b2c3d4e5f6g7h8i9j0k1l2m3n4",
			expected: "gh********************n4",
		},
		{
			name:     "Stripe key (24 chars)",
			key:      "STRIPE_KEY",
			value:    "sk_live_abcdefghijklmnop",
			expected: "sk********************op",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Mask(tt.key, tt.value)
			if got != tt.expected {
				t.Errorf("Mask(%q, %q) = %q, want %q", tt.key, tt.value, got, tt.expected)
			}
		})
	}
}

func TestMaskIfSensitive(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		value          string
		shouldBeMasked bool
	}{
		// Should mask based on key
		{
			name:           "password key",
			key:            "DATABASE_PASSWORD",
			value:          "MySecretPassword123",
			shouldBeMasked: true,
		},
		{
			name:           "API key",
			key:            "API_KEY",
			value:          "abc123xyz789",
			shouldBeMasked: true,
		},
		{
			name:           "secret token",
			key:            "SECRET_TOKEN",
			value:          "token_value_here",
			shouldBeMasked: true,
		},

		// Should mask based on value prefix
		{
			name:           "Stripe key (non-sensitive key name)",
			key:            "PAYMENT_CONFIG",
			value:          "sk_live_abc123xyz789",
			shouldBeMasked: true,
		},
		{
			name:           "GitHub token (generic key name)",
			key:            "TOKEN",
			value:          "ghp_a1b2c3d4e5f6g7h8i9j0",
			shouldBeMasked: true,
		},
		{
			name:           "AWS key (generic key name)",
			key:            "AWS",
			value:          "AKIAIOSFODNN7EXAMPLE",
			shouldBeMasked: true,
		},

		// Should mask based on value pattern (JWT)
		{
			name:           "JWT token (non-sensitive key)",
			key:            "AUTH",
			value:          "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
			shouldBeMasked: true,
		},

		// Should mask based on value pattern (base64)
		{
			name:           "Base64 secret (generic key)",
			key:            "CONFIG",
			value:          "YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXoxMjM0NTY3ODkw",
			shouldBeMasked: true,
		},

		// Should mask based on value pattern (hex)
		{
			name:           "Hex secret (generic key)",
			key:            "VALUE",
			value:          "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4",
			shouldBeMasked: true,
		},

		// Should NOT mask
		{
			name:           "database URL",
			key:            "DATABASE_URL",
			value:          "postgres://localhost:5432/mydb",
			shouldBeMasked: false,
		},
		{
			name:           "port number",
			key:            "PORT",
			value:          "3000",
			shouldBeMasked: false,
		},
		{
			name:           "debug flag",
			key:            "DEBUG",
			value:          "true",
			shouldBeMasked: false,
		},
		{
			name:           "app name",
			key:            "APP_NAME",
			value:          "MyApplication",
			shouldBeMasked: false,
		},
		{
			name:           "log level",
			key:            "LOG_LEVEL",
			value:          "info",
			shouldBeMasked: false,
		},
		{
			name:           "URL",
			key:            "API_URL",
			value:          "https://api.example.com",
			shouldBeMasked: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaskIfSensitive(tt.key, tt.value)

			isMasked := got != tt.value
			if isMasked != tt.shouldBeMasked {
				t.Errorf("MaskIfSensitive(%q, %q) masked=%v, want masked=%v\nGot: %q",
					tt.key, tt.value, isMasked, tt.shouldBeMasked, got)
			}

			// If it should be masked, verify it contains asterisks or is "(empty)" or "***"
			if tt.shouldBeMasked && !strings.Contains(got, "*") {
				t.Errorf("Expected masked value to contain *, got: %q", got)
			}
		})
	}
}

func TestMaskMap(t *testing.T) {
	input := map[string]string{
		"DATABASE_URL":      "postgres://localhost:5432/mydb",
		"API_KEY":           "sk_live_abc123xyz789",
		"DATABASE_PASSWORD": "MySecretPassword123",
		"PORT":              "3000",
		"DEBUG":             "true",
		"SECRET_TOKEN":      "ghp_a1b2c3d4e5f6g7h8i9j0",
		"APP_NAME":          "EnvQuack",
		"GITHUB_TOKEN":      "ghp_xyz123abc456def789",
		"LOG_LEVEL":         "info",
	}

	result := MaskMap(input)

	// Verify we got a new map with same number of keys
	if len(result) != len(input) {
		t.Errorf("MaskMap returned %d keys, want %d", len(result), len(input))
	}

	// Non-sensitive values should be unchanged
	nonSensitive := []string{"DATABASE_URL", "PORT", "DEBUG", "APP_NAME", "LOG_LEVEL"}
	for _, key := range nonSensitive {
		if result[key] != input[key] {
			t.Errorf("%s should not be masked: got %q, want %q", key, result[key], input[key])
		}
	}

	// Sensitive values should be masked
	sensitive := []string{"API_KEY", "DATABASE_PASSWORD", "SECRET_TOKEN", "GITHUB_TOKEN"}
	for _, key := range sensitive {
		if result[key] == input[key] {
			t.Errorf("%s should be masked but wasn't", key)
		}
		if !strings.Contains(result[key], "*") {
			t.Errorf("Masked %s should contain *, got: %q", key, result[key])
		}
	}
}

func TestSetMaskingEnabled(t *testing.T) {
	// Save original state
	originalState := disableMasking
	defer func() { disableMasking = originalState }()

	// Test disabling masking
	SetMaskingEnabled(false)
	result := MaskIfSensitive("PASSWORD", "secret123")
	if result != "secret123" {
		t.Errorf("When masking disabled, expected unmasked value, got: %q", result)
	}

	// Test enabling masking
	SetMaskingEnabled(true)
	result = MaskIfSensitive("PASSWORD", "secret123")
	if result == "secret123" {
		t.Error("When masking enabled, expected masked value")
	}
	if !strings.Contains(result, "*") {
		t.Errorf("Masked value should contain *, got: %q", result)
	}
}

func TestMaskingEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{name: "unicode characters", key: "PASSWORD", value: "p@ssw0rd™"},
		{name: "special characters", key: "SECRET", value: "!@#$%^&*()"},
		{name: "spaces", key: "API_KEY", value: "key with spaces"},
		{name: "newlines", key: "TOKEN", value: "token\nwith\nnewlines"},
		{name: "tabs", key: "CREDENTIAL", value: "cred\twith\ttabs"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskIfSensitive(tt.key, tt.value)
			// Should mask these (they have sensitive keys)
			if result == tt.value {
				t.Error("Expected value to be masked")
			}
			// Should contain asterisks (unless very short)
			if len(tt.value) > 4 && !strings.Contains(result, "*") {
				t.Errorf("Expected masked value to contain *, got: %q", result)
			}
		})
	}
}

func TestRealWorldExamples(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		value          string
		shouldBeMasked bool
		description    string
	}{
		{
			name:           "Stripe live secret",
			key:            "STRIPE_SECRET_KEY",
			value:          "sk_live_51H7xLmF9Y3K4p5N6c7D8e9F0g1H2i3J4k5L6m7N8o9P0q1R2s",
			shouldBeMasked: true,
			description:    "Stripe secret keys should always be masked",
		},
		{
			name:           "GitHub PAT",
			key:            "GITHUB_TOKEN",
			value:          "ghp_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8",
			shouldBeMasked: true,
			description:    "GitHub personal access tokens should be masked",
		},
		{
			name:           "AWS access key",
			key:            "AWS_ACCESS_KEY_ID",
			value:          "AKIAIOSFODNN7EXAMPLE",
			shouldBeMasked: true,
			description:    "AWS access keys should be masked",
		},
		{
			name:           "Database password",
			key:            "DB_PASSWORD",
			value:          "MyS3cr3tP@ssw0rd!",
			shouldBeMasked: true,
			description:    "Database passwords should be masked",
		},
		{
			name:           "JWT token",
			key:            "JWT_TOKEN",
			value:          "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			shouldBeMasked: true,
			description:    "JWT tokens should be masked",
		},
		{
			name:           "Redis URL with password",
			key:            "REDIS_URL",
			value:          "redis://user:password123@localhost:6379",
			shouldBeMasked: false,
			description:    "URLs should not be masked (even with embedded passwords)",
		},
		{
			name:           "Public API endpoint",
			key:            "API_ENDPOINT",
			value:          "https://api.example.com/v1",
			shouldBeMasked: false,
			description:    "Public API endpoints should not be masked",
		},
		{
			name:           "Environment name",
			key:            "ENVIRONMENT",
			value:          "production",
			shouldBeMasked: false,
			description:    "Environment names should not be masked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskIfSensitive(tt.key, tt.value)
			isMasked := result != tt.value

			if isMasked != tt.shouldBeMasked {
				t.Errorf("%s: masked=%v, want=%v\n  Value: %q\n  Result: %q",
					tt.description, isMasked, tt.shouldBeMasked, tt.value, result)
			}
		})
	}
}
