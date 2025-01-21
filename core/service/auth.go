package service

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"log"
	"messanger/config"
	"messanger/core/ports"
	"messanger/models"
	tr "messanger/pkg/error_trace"
	"time"
)

type AuthService struct {
	cache ports.TokenCache
	repo  ports.AuthUsers
	cfg   *config.AuthServiceConfig
}

var ErrUnauthorized = errors.New("unauthorized")

func NewAuthService(cache ports.TokenCache, repo ports.AuthUsers, cfg *config.AuthServiceConfig) *AuthService {
	cache.SetTTL(cfg.RefreshTokenTTL)
	return &AuthService{
		cache: cache,
		repo:  repo,
		cfg:   cfg,
	}
}

func (s *AuthService) CreateUser(user *models.AuthUser) error {
	exist, err := s.repo.IsExist(user.Email)
	if err != nil {
		return tr.Trace(err)
	}
	if exist {
		return errors.New("user with this email already exists")
	}

	user.Password = s.hashPassword(user.Password)
	if err := s.repo.New(user); err != nil {
		return tr.Trace(err)
	}
	return nil
}

func (s *AuthService) Login(email, password string) (*models.Tokens, error) {
	ok, err := s.repo.CheckPassword(s.hashPassword(password), email)
	if err != nil {
		return nil, tr.Trace(err)
	}
	if !ok {
		return nil, ErrUnauthorized
	}
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return nil, tr.Trace(err)
	}

	access, accessExpires, err := s.newAccessToken(user.Id)
	if err != nil {
		return nil, tr.Trace(err)
	}

	refreshExpired := time.Now().Add(s.cfg.RefreshTokenTTL)
	refresh := s.newRefreshToken()
	if err := s.cache.Set(refresh, user.Id); err != nil {
		return nil, tr.Trace(err)
	}

	return &models.Tokens{
		AccessToken:           access,
		RefreshToken:          refresh,
		AccessTokenExpiresAt:  accessExpires,
		RefreshTokenExpiresAt: refreshExpired,
	}, nil
}

func (s *AuthService) UpdateTokens(refresh string) (*models.Tokens, error) {
	userId, err := s.cache.Get(refresh)
	if err != nil {
		return nil, tr.Trace(err)
	}
	if userId == 0 {
		return nil, ErrUnauthorized
	}

	if err := s.cache.Del(refresh); err != nil {
		return nil, tr.Trace(err)
	}

	access, accessExpired, err := s.newAccessToken(userId)
	if err != nil {
		return nil, tr.Trace(err)
	}

	refreshExpired := time.Now().Add(s.cfg.RefreshTokenTTL)
	newRefresh := s.newRefreshToken()
	if err := s.cache.Set(newRefresh, userId); err != nil {
		return nil, tr.Trace(err)
	}

	return &models.Tokens{
		AccessToken:           access,
		RefreshToken:          refresh,
		AccessTokenExpiresAt:  accessExpired,
		RefreshTokenExpiresAt: refreshExpired,
	}, err
}

var tokenKey = []byte("ac4873yqc34v5")

func (s *AuthService) CheckAccessToken(access string) (int, error) {
	claims, err := jwt.Parse(access, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return tokenKey, nil
	})
	if err != nil {
		return 0, tr.Trace(err)
	}
	mapClaims, ok := claims.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("claims is not a map")
	}

	userId, ok := mapClaims["id"]
	if !ok {
		return 0, errors.New("userId is not in map claims")
	}
	tUnix, ok := mapClaims["expires"]
	if !ok {
		return 0, errors.New("expires time is not in map claims")
	}

	log.Println(
		time.Unix(int64(tUnix.(float64)), 0),
	)

	if time.Now().After(time.Unix(int64(tUnix.(float64)), 0)) {
		return 0, ErrUnauthorized
	}

	return int(userId.(float64)), nil
}

func (s *AuthService) newAccessToken(userId int) (string, time.Time, error) {
	expires := time.Now().Add(s.cfg.AccessTokenTTL)

	claims := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"id":      userId,
		"expires": expires.Unix(),
	})
	token, err := claims.SignedString(tokenKey)
	if err != nil {
		return "", expires, tr.Trace(err)
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
