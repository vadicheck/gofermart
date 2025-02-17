package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID int `json:"user_id"`
}

func BuildJWTString(jwtSecret string, jwtTokenExpire time.Duration, userID int) (tokenString string, err error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(jwtTokenExpire)),
		},
		UserID: userID,
	})

	tokenString, err = token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", fmt.Errorf("can't sign token: %w", err)
	}

	return tokenString, nil
}

func DecodeJwtToken(jwtToken, jwtSecret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(jwtToken, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("can't parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is invalid")
	}

	return token.Claims.(*Claims), nil
}
