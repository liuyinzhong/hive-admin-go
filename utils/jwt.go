package utils

import (
	"errors"
	"hive-admin-go/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret []byte

func InitJWT() {
	jwtSecret = []byte(config.AppConfig.JWT.Secret)
}

type Claims struct {
	UserID string `json:"userId"`
	jwt.RegisteredClaims
}

func GenerateToken(userID string) (string, error) {
	expireTime := time.Now().Add(time.Duration(config.AppConfig.JWT.Expire) * time.Hour)
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func IsTokenBlacklisted(tokenString string) bool {
	return false
}

func AddTokenToBlacklist(tokenString string) {
}

func ValidateToken(tokenString string) bool {
	if IsTokenBlacklisted(tokenString) {
		return false
	}

	_, err := ParseToken(tokenString)
	return err == nil
}
