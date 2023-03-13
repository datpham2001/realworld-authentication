package helper

import (
	"errors"
	"realworld-authentication/config"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type SignedDetails struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	UserID   string `json:"userId"`
	jwt.StandardClaims
}

func GenerateJWT(email, username, userId string) (map[string]string, error) {
	// gen access token
	expirationTimeAT := time.Now().Add(1 * time.Hour)
	claimsAT := &SignedDetails{
		Username: username,
		Email:    email,
		UserID:   userId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTimeAT.Unix(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claimsAT)
	accessTokenString, err := accessToken.SignedString(config.AppConfig.TokenSecret)
	if err != nil {
		return nil, err
	}

	// gen refresh token
	expirationTimeRT := time.Now().Add(24 * time.Hour)
	claimsRT := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTimeRT.Unix(),
		},
	}

	refresToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claimsRT)
	refresTokenString, err := refresToken.SignedString(config.AppConfig.TokenSecret)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"accessToken":  accessTokenString,
		"refreshToken": refresTokenString,
	}, nil
}

func ValidateToken(signedToken string) (*SignedDetails, error) {
	token, err := jwt.ParseWithClaims(signedToken, &SignedDetails{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(config.AppConfig.TokenSecret), nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*SignedDetails)
	if !ok || !token.Valid {
		return nil, errors.New("token in invalid")
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		return nil, errors.New("token is expired")
	}

	return claims, nil
}
