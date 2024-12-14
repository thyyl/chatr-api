package common

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var (
	JwtSecret           string
	JwtExpirationSecond int64
)

var (
	ErrorInvalidToken = errors.New("invalid token")
	ErrorTokenExpired = errors.New("token expired")
)

type JWTClaims struct {
	ChannelId uint64
	jwt.RegisteredClaims
}

type AuthPayload struct {
	AccessToken string
}

type AuthResponse struct {
	ChannelId uint64
	Expired   bool
}

func Auth(authPayload *AuthPayload) (*AuthResponse, error) {
	token, err := parseToken(authPayload.AccessToken)
	if err != nil {
		v := err.(*jwt.ValidationError)
		if v.Errors == jwt.ValidationErrorExpired {
			return &AuthResponse{
				Expired: true,
			}, nil
		}
		return nil, ErrorInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !(ok && token.Valid) {
		return nil, ErrorInvalidToken
	}

	return &AuthResponse{
		ChannelId: claims.ChannelId,
		Expired:   false,
	}, nil
}

func NewJWT(channelId uint64) (string, error) {
	expiresAt := time.Now().Add(time.Duration(JwtExpirationSecond) * time.Second)
	jwtClaims := &JWTClaims{
		ChannelId: channelId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	accessToken, err := token.SignedString([]byte(JwtSecret))
	if err != nil {
		return "", err
	}
	return accessToken, nil
}

func parseToken(accessToken string) (*jwt.Token, error) {
	return jwt.ParseWithClaims(accessToken, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(JwtSecret), nil
	})
}
