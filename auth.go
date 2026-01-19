// Package libecto provides a client library for the Ghost Admin API.
//
// It handles JWT authentication, API requests, and provides typed responses
// for all Ghost Admin API endpoints including posts, pages, tags, users,
// newsletters, webhooks, and images.
package libecto

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ParseAPIKey splits a Ghost Admin API key into id and secret.
// Ghost API keys have the format "{id}:{secret}" where secret is hex-encoded.
// It returns an error if the format is invalid or the secret is not valid hex.
func ParseAPIKey(apiKey string) (id string, secret []byte, err error) {
	parts := strings.Split(apiKey, ":")
	if len(parts) != 2 {
		return "", nil, fmt.Errorf("invalid API key format: expected 'id:secret'")
	}

	id = parts[0]
	if id == "" {
		return "", nil, fmt.Errorf("invalid API key format: id cannot be empty")
	}

	if parts[1] == "" {
		return "", nil, fmt.Errorf("invalid API key format: secret cannot be empty")
	}

	secret, err = hex.DecodeString(parts[1])
	if err != nil {
		return "", nil, fmt.Errorf("invalid API key secret: %w", err)
	}
	return id, secret, nil
}

// GenerateToken creates a JWT token for Ghost Admin API authentication.
// The token is signed with HS256 and is valid for 5 minutes.
// It returns an error if the API key is invalid.
func GenerateToken(apiKey string) (string, error) {
	id, secret, err := ParseAPIKey(apiKey)
	if err != nil {
		return "", err
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"iat": now.Unix(),
		"exp": now.Add(5 * time.Minute).Unix(),
		"aud": "/admin/",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["kid"] = id
	token.Header["typ"] = "JWT"

	return token.SignedString(secret)
}

// GenerateTokenWithTime creates a JWT token with a specific timestamp.
// This is useful for testing or when a specific token time is needed.
// The token is valid for 5 minutes from the given time.
func GenerateTokenWithTime(apiKey string, now time.Time) (string, error) {
	id, secret, err := ParseAPIKey(apiKey)
	if err != nil {
		return "", err
	}

	claims := jwt.MapClaims{
		"iat": now.Unix(),
		"exp": now.Add(5 * time.Minute).Unix(),
		"aud": "/admin/",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["kid"] = id
	token.Header["typ"] = "JWT"

	return token.SignedString(secret)
}
