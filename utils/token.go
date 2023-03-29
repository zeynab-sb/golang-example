package utils

import (
	"errors"
	"fmt"
	"golang-example/config"
	"time"

	"github.com/golang-jwt/jwt"
)

type jwtClaim struct {
	ID uint `json:"id"`
	jwt.StandardClaims
}

func GenerateToken(id uint) (string, error) {
	claims := &jwtClaim{
		ID: id,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(config.C.Token.ExpiresIn).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.C.Token.Secret))

	if err != nil {
		return "", fmt.Errorf("generating JWT Token failed: %w", err)
	}

	return tokenString, nil
}

func ValidateToken(signedToken string) (uint, error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&jwtClaim{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(config.C.Token.Secret), nil
		},
	)

	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(*jwtClaim)
	if !ok {
		return 0, errors.New("couldn't parse claims")
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		return 0, errors.New("token expired")
	}

	return claims.ID, nil
}
