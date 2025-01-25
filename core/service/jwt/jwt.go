package jwt

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"messanger/pkg/errors"
	"net/http"
	"time"
)

const ErrInvalidToken = "invalid token"

type TokenManager struct {
	TTL     time.Duration
	signKey []byte
}

func NewTokenManager(ttl time.Duration, signKey []byte) *TokenManager {
	return &TokenManager{
		TTL:     ttl,
		signKey: signKey,
	}
}

func (m *TokenManager) DecodeToken(access string) (int, *errors.Error) {
	claims, err := jwt.Parse(access, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return m.signKey, nil
	})
	if err != nil {
		return 0, errors.New(err, ErrInvalidToken, http.StatusUnauthorized)
	}
	mapClaims, ok := claims.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("claims is not a map", ErrInvalidToken, http.StatusUnauthorized)
	}

	id, ok := mapClaims["id"]
	if !ok {
		return 0, errors.New("id is not in map claims", ErrInvalidToken, http.StatusUnauthorized)
	}
	tUnix, ok := mapClaims["expires"]
	if !ok {
		return 0, errors.New("expires time is not in map claims", ErrInvalidToken, http.StatusUnauthorized)
	}

	if time.Now().After(time.Unix(int64(tUnix.(float64)), 0)) {
		return 0, errors.New1Msg("token expired", http.StatusUnauthorized)
	}

	return int(id.(float64)), nil
}

func (m *TokenManager) NewToken(id int) (string, time.Time, *errors.Error) {
	expires := time.Now().Add(m.TTL)

	claims := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"id":      id,
		"expires": expires.Unix(),
	})
	token, err := claims.SignedString(m.signKey)
	if err != nil {
		return "", expires, errors.New(err, "create token error", http.StatusBadRequest)
	}
	return token, expires, nil
}
