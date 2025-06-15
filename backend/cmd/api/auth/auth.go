package auth

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
)

// ValidateToken validates the provided JWT token and extracts the user_id claim.
func ValidateToken(tokenStr string) (float64, error) {
	// Parse JWT token
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// Ensure the token is signed with the correct signing method (HMAC, RSA, etc.)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		// Return the secret key to verify the token
		return []byte("your-secret-key"), nil
	})
	if err != nil || !token.Valid {
		return 0, errors.New("invalid token")
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid token claims")
	}

	// Extract user ID from the claims
	userID, ok := claims["user_id"].(float64)
	if !ok {
		return 0, errors.New("invalid user ID")
	}

	return userID, nil
}
