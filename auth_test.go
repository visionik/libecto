package libecto

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Valid test API key: id is "test123", secret is hex-encoded "secretkey123"
const (
	validAPIKey     = "test123:7365637265746b6579313233"
	validAPIKeyID   = "test123"
	validAPISecret  = "secretkey123"
	invalidHexKey   = "test123:notvalidhex!!"
	malformedKey    = "no-colon-here"
	emptyIDKey      = ":7365637265746b6579313233"
	emptySecretKey  = "test123:"
	multiColonKey   = "test:123:456"
)

func TestParseAPIKey(t *testing.T) {
	tests := []struct {
		name       string
		apiKey     string
		wantID     string
		wantSecret string
		wantErr    bool
		errContain string
	}{
		{
			name:       "valid API key",
			apiKey:     validAPIKey,
			wantID:     validAPIKeyID,
			wantSecret: validAPISecret,
			wantErr:    false,
		},
		{
			name:       "missing colon separator",
			apiKey:     malformedKey,
			wantErr:    true,
			errContain: "expected 'id:secret'",
		},
		{
			name:       "empty ID",
			apiKey:     emptyIDKey,
			wantErr:    true,
			errContain: "id cannot be empty",
		},
		{
			name:       "empty secret",
			apiKey:     emptySecretKey,
			wantErr:    true,
			errContain: "secret cannot be empty",
		},
		{
			name:       "invalid hex in secret",
			apiKey:     invalidHexKey,
			wantErr:    true,
			errContain: "invalid API key secret",
		},
		{
			name:       "multiple colons",
			apiKey:     multiColonKey,
			wantErr:    true,
			errContain: "expected 'id:secret'",
		},
		{
			name:    "empty string",
			apiKey:  "",
			wantErr: true,
		},
		{
			name:       "only colon",
			apiKey:     ":",
			wantErr:    true,
			errContain: "id cannot be empty",
		},
		{
			name:       "whitespace in key",
			apiKey:     "test 123:7365637265746b6579313233",
			wantID:     "test 123",
			wantSecret: validAPISecret,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, secret, err := ParseAPIKey(tt.apiKey)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContain != "" {
					assert.Contains(t, err.Error(), tt.errContain)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantID, id)
			assert.Equal(t, tt.wantSecret, string(secret))
		})
	}
}

func TestGenerateToken(t *testing.T) {
	tests := []struct {
		name    string
		apiKey  string
		wantErr bool
	}{
		{
			name:    "valid API key",
			apiKey:  validAPIKey,
			wantErr: false,
		},
		{
			name:    "invalid API key format",
			apiKey:  malformedKey,
			wantErr: true,
		},
		{
			name:    "invalid hex secret",
			apiKey:  invalidHexKey,
			wantErr: true,
		},
		{
			name:    "empty API key",
			apiKey:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateToken(tt.apiKey)
			if tt.wantErr {
				require.Error(t, err)
				assert.Empty(t, token)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, token)

			// Verify token structure
			parts := strings.Split(token, ".")
			assert.Len(t, parts, 3, "JWT should have 3 parts")
		})
	}
}

func TestGenerateTokenWithTime(t *testing.T) {
	fixedTime := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		apiKey  string
		time    time.Time
		wantErr bool
	}{
		{
			name:    "valid with fixed time",
			apiKey:  validAPIKey,
			time:    fixedTime,
			wantErr: false,
		},
		{
			name:    "invalid API key",
			apiKey:  malformedKey,
			time:    fixedTime,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateTokenWithTime(tt.apiKey, tt.time)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, token)
		})
	}
}

func TestGenerateToken_ValidJWT(t *testing.T) {
	token, err := GenerateToken(validAPIKey)
	require.NoError(t, err)

	// Parse without validation to check claims
	parsed, _, err := jwt.NewParser().ParseUnverified(token, jwt.MapClaims{})
	require.NoError(t, err)

	// Check header
	assert.Equal(t, "JWT", parsed.Header["typ"])
	assert.Equal(t, validAPIKeyID, parsed.Header["kid"])
	assert.Equal(t, "HS256", parsed.Header["alg"])

	// Check claims
	claims, ok := parsed.Claims.(jwt.MapClaims)
	require.True(t, ok)

	assert.Equal(t, "/admin/", claims["aud"])
	assert.NotNil(t, claims["iat"])
	assert.NotNil(t, claims["exp"])

	// Verify exp is 5 minutes after iat
	iat := int64(claims["iat"].(float64))
	exp := int64(claims["exp"].(float64))
	assert.Equal(t, int64(300), exp-iat, "Token should be valid for 5 minutes")
}

func TestGenerateToken_Verifiable(t *testing.T) {
	token, err := GenerateToken(validAPIKey)
	require.NoError(t, err)

	// Verify the token with the secret
	_, secret, _ := ParseAPIKey(validAPIKey)
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	require.NoError(t, err)
	assert.True(t, parsed.Valid)
}

func TestGenerateTokenWithTime_Deterministic(t *testing.T) {
	fixedTime := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)

	token1, err := GenerateTokenWithTime(validAPIKey, fixedTime)
	require.NoError(t, err)

	token2, err := GenerateTokenWithTime(validAPIKey, fixedTime)
	require.NoError(t, err)

	// Same time should produce same token
	assert.Equal(t, token1, token2)
}

// Fuzz tests for security-sensitive parsing

func FuzzParseAPIKey(f *testing.F) {
	// Seed corpus
	f.Add("test123:7365637265746b6579313233")
	f.Add("no-colon")
	f.Add(":")
	f.Add(":secret")
	f.Add("id:")
	f.Add("test:123:456")
	f.Add("")
	f.Add("a:b")
	f.Add("abc:def")
	f.Add("id:0123456789abcdef")
	f.Add("id:ABCDEF")
	f.Add("id:ghijkl") // invalid hex
	f.Add("id:!@#$%^")
	f.Add("id:00")
	f.Add("id:ff")
	f.Add("test\x00id:abc")
	f.Add("id\ntest:abc")
	f.Add(" :abc")
	f.Add("test : abc")
	f.Add("\t:\t")
	f.Add(strings.Repeat("a", 1000) + ":" + strings.Repeat("0", 1000))
	// Add more for fuzzing coverage
	for i := 0; i < 30; i++ {
		f.Add("test" + string(rune(i+48)) + ":" + strings.Repeat("ab", i+1))
	}

	f.Fuzz(func(t *testing.T, apiKey string) {
		// Should not panic
		_, _, _ = ParseAPIKey(apiKey)
	})
}

// Benchmark tests

func BenchmarkParseAPIKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, _ = ParseAPIKey(validAPIKey)
	}
}

func BenchmarkGenerateToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = GenerateToken(validAPIKey)
	}
}
