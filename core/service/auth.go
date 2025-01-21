package service

import (
	"crypto/sha256"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"log"
	"messanger/config"
	"messanger/core/ports"
	"messanger/domain"
	"messanger/pkg/errors"
	"net/http"
	"time"
)

type AuthService struct {
	cache ports.TokenCache
	repo  ports.UsersRepo
	cfg   *config.AuthServiceConfig
}

func NewAuthService(cache ports.TokenCache, repo ports.UsersRepo, cfg *config.AuthServiceConfig) *AuthService {
	cache.SetTTL(cfg.RefreshTokenTTL)
	return &AuthService{
		cache: cache,
		repo:  repo,
		cfg:   cfg,
	}
}

func (s *AuthService) Login(email, password string) (*domain.Tokens, *errors.Error) {
	user, err := s.repo.GetInfoByEmail(email)
	if err != nil {
		return nil, err.Trace()
	}
	if user == nil {
		return nil, errors.New(domain.ErrUserAlreadyExists, domain.ErrUserAlreadyExists, http.StatusNotFound)
	}

	if user.Password != s.hashPassword(password) {
		return nil, errors.New1Msg("invalid password", http.StatusUnauthorized)
	}

	access, accessExpires, err := s.newAccessToken(user.Id)
	if err != nil {
		return nil, err.Trace()
	}

	refreshExpired := time.Now().Add(s.cfg.RefreshTokenTTL)
	refresh := s.newRefreshToken()
	if err := s.cache.Set(refresh, user.Id); err != nil {
		return nil, err.Trace()
	}

	return &domain.Tokens{
		AccessToken:           access,
		RefreshToken:          refresh,
		AccessTokenExpiresAt:  accessExpires,
		RefreshTokenExpiresAt: refreshExpired,
	}, nil
}

func (s *AuthService) UpdateTokens(refresh string) (*domain.Tokens, *errors.Error) {
	userId, err := s.cache.Get(refresh)
	if err != nil {
		return nil, err.Trace()
	}
	if userId == 0 {
		return nil, errors.New(domain.ErrUnauthorized, domain.ErrUnauthorized, http.StatusUnauthorized)
	}

	if err := s.cache.Del(refresh); err != nil {
		return nil, err.Trace()
	}

	access, accessExpired, err := s.newAccessToken(userId)
	if err != nil {
		return nil, err.Trace()
	}

	refreshExpired := time.Now().Add(s.cfg.RefreshTokenTTL)
	newRefresh := s.newRefreshToken()
	if err := s.cache.Set(newRefresh, userId); err != nil {
		return nil, err.Trace()
	}

	return &domain.Tokens{
		AccessToken:           access,
		RefreshToken:          refresh,
		AccessTokenExpiresAt:  accessExpired,
		RefreshTokenExpiresAt: refreshExpired,
	}, err
}

var tokenKey = []byte("ac4873yqc34v5")

const (
	ErrInvalidToken = "invalid token"
)

func (s *AuthService) CheckAccessToken(access string) (int, *errors.Error) {
	claims, err := jwt.Parse(access, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return tokenKey, nil
	})
	if err != nil {
		return 0, errors.New(err, ErrInvalidToken, http.StatusUnauthorized)
	}
	mapClaims, ok := claims.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("claims is not a map", ErrInvalidToken, http.StatusUnauthorized)
	}

	userId, ok := mapClaims["id"]
	if !ok {
		return 0, errors.New("userId is not in map claims", ErrInvalidToken, http.StatusUnauthorized)
	}
	tUnix, ok := mapClaims["expires"]
	if !ok {
		return 0, errors.New("expires time is not in map claims", ErrInvalidToken, http.StatusUnauthorized)
	}

	log.Println(
		time.Unix(int64(tUnix.(float64)), 0),
	)

	if time.Now().After(time.Unix(int64(tUnix.(float64)), 0)) {
		return 0, errors.New1Msg("token expired", http.StatusUnauthorized)
	}

	return int(userId.(float64)), nil
}

func (s *AuthService) newAccessToken(userId int) (string, time.Time, *errors.Error) {
	expires := time.Now().Add(s.cfg.AccessTokenTTL)

	claims := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"id":      userId,
		"expires": expires.Unix(),
	})
	token, err := claims.SignedString(tokenKey)
	if err != nil {
		return "", expires, errors.New(err, "invalid token", http.StatusBadRequest)
	}
	return token, expires, nil
}

func (s *AuthService) newRefreshToken() string {
	uid, _ := uuid.NewUUID()
	return uid.String()
}

var passwordSalt = []byte("avm84ut397q58y")

func (s *AuthService) hashPassword(password string) string {
	h := sha256.New()
	h.Write([]byte(password))
	return string(h.Sum(passwordSalt))
}
