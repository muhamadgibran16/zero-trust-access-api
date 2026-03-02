package security

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTClaims represents the claims in a JWT token
type JWTClaims struct {
	UserID uuid.UUID `json:"userID"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken generates a new JWT token for a given user
func GenerateToken(userID uuid.UUID, role string, secret string, expireHours int) (string, error) {
	claims := &JWTClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// DeviceTokenClaims represents the claims in a device JWT token
type DeviceTokenClaims struct {
	DeviceID   string `json:"deviceID"`
	MacAddress string `json:"macAddress"`
	CertThumb  string `json:"certThumb"`
	jwt.RegisteredClaims
}

// GenerateDeviceToken generates a long-lived JWT token for an approved device
func GenerateDeviceToken(deviceID, macAddress, certThumb, secret string) (string, error) {
	claims := &DeviceTokenClaims{
		DeviceID:   deviceID,
		MacAddress: macAddress,
		CertThumb:  certThumb,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(365 * 24 * time.Hour)), // 1 year expiry for hardware tokens
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "device-identity",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateDeviceToken validates a device JWT token
func ValidateDeviceToken(tokenString, secret string) (*DeviceTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &DeviceTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*DeviceTokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid device token")
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString string, secret string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
