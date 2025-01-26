package service

import (
	"github.com/google/uuid"
	"messanger/config"
	"messanger/core/ports"
	"messanger/core/service/jwt"
	"messanger/domain"
	"messanger/pkg/errors"
	"net/http"
	"time"
)

type AuthService struct {
	cache           ports.Cache
	repo            ports.UsersRepo
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	TokenManager    *jwt.TokenManager
}

func NewAuthService(cache ports.Cache, repo ports.UsersRepo, cfg *config.AuthServiceConfig) *AuthService {
	accessTokenTTL := time.Duration(cfg.AccessTokenTTLMin) * time.Minute
	refreshTokenTTL := time.Duration(cfg.RefreshTokenTTLDays) * time.Hour * 24

	TokenManager := jwt.NewTokenManager(time.Duration(cfg.AccessTokenTTLMin)*time.Minute, []byte(cfg.AccessTokenSignKey))

	return &AuthService{
		cache:           cache,
		repo:            repo,
		refreshTokenTTL: refreshTokenTTL,
		accessTokenTTL:  accessTokenTTL,
		TokenManager:    TokenManager,
	}
}

func (s *AuthService) Login(email, password string) (*domain.Tokens, *errors.Error) {
	user, err := s.repo.GetByEmailWithPass(email, password)
	if err != nil {
		return nil, err.Trace()
	}

	access, accessExpires, err := s.TokenManager.NewToken(user.Id)
	if err != nil {
		return nil, err.Trace()
	}

	refreshExpired := time.Now().Add(s.refreshTokenTTL)
	refresh := s.newRefreshToken()
	if err := s.cache.Set(refresh, user.Id, s.refreshTokenTTL); err != nil {
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
		return nil, errors.New1Msg("missing user id", http.StatusUnauthorized)
	}

	if err := s.cache.Del(refresh); err != nil {
		return nil, err.Trace()
	}

	access, accessExpired, err := s.TokenManager.NewToken(userId)
	if err != nil {
		return nil, err.Trace()
	}

	refreshExpired := time.Now().Add(s.refreshTokenTTL)
	newRefresh := s.newRefreshToken()
	if err := s.cache.Set(newRefresh, userId, s.refreshTokenTTL); err != nil {
		return nil, err.Trace()
	}

	return &domain.Tokens{
		AccessToken:           access,
		RefreshToken:          refresh,
		AccessTokenExpiresAt:  accessExpired,
		RefreshTokenExpiresAt: refreshExpired,
	}, err
}

func (s *AuthService) newRefreshToken() string {
	uid, _ := uuid.NewUUID()
	return uid.String()
}
