package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT claims with custom fields
type Claims struct {
	jwt.RegisteredClaims
	Login  string `json:"login"`
	UserID int64  `json:"user_id"`
}

// JWTManager handles JWT token generation and validation
type JWTManager struct {
	secret []byte
	maxAge time.Duration
}

// NewJWTManager creates a new JWT manager with the given secret and max age
func NewJWTManager(secret string, maxAge time.Duration) *JWTManager {
	return &JWTManager{
		secret: []byte(secret),
		maxAge: maxAge,
	}
}

// GenerateToken creates a new JWT token for the given user
func (m *JWTManager) GenerateToken(username, login string, userID int64) (string, error) {
	now := time.Now()
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   username,
			ExpiresAt: jwt.NewNumericDate(now.Add(m.maxAge)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "fw-app",
		},
		Login:  login,
		UserID: userID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
