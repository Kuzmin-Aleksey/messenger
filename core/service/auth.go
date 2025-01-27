package service

import (
	"fmt"
	"github.com/google/uuid"
	"log"
	"messanger/config"
	"messanger/core/ports"
	"messanger/core/service/jwt"
	"messanger/domain"
	"messanger/pkg/errors"
	"net/http"
	"time"
)

type AuthService struct {
	cache            ports.Cache
	repo             ports.UsersRepo
	accessTokenTTL   time.Duration
	refreshTokenTTL  time.Duration
	TokenManager     *jwt.TokenManager
	blockDuration    time.Duration
	maxLoginAttempts int
}

func NewAuthService(cache ports.Cache, repo ports.UsersRepo, cfg *config.AuthServiceConfig) *AuthService {
	accessTokenTTL := time.Duration(cfg.AccessTokenTTLMin) * time.Minute
	refreshTokenTTL := time.Duration(cfg.RefreshTokenTTLDays) * time.Hour * 24
	blockDuration := time.Duration(cfg.DurationBlockUserMin) * time.Minute

	TokenManager := jwt.NewTokenManager(time.Duration(cfg.AccessTokenTTLMin)*time.Minute, []byte(cfg.AccessTokenSignKey))

	log.Println(blockDuration, cfg.LoginAttempts)
	return &AuthService{
		cache:            cache,
		repo:             repo,
		refreshTokenTTL:  refreshTokenTTL,
		accessTokenTTL:   accessTokenTTL,
		TokenManager:     TokenManager,
		blockDuration:    blockDuration,
		maxLoginAttempts: cfg.LoginAttempts,
	}
}

func (s *AuthService) Login(email, password string) (*domain.Tokens, *errors.Error) {
	blocked, unblockTime, err := s.isBlocked(email)
	if err != nil {
		return nil, err.Trace()
	}
	if blocked {
		return nil, errors.New(fmt.Sprintf("user %s blocked", email),
			fmt.Sprintf("exceeded number of login attempts, next attempt - %s", unblockTime.Format("01.02 15:04")),
			http.StatusTooManyRequests)
	}

	user, err := s.repo.GetByEmailWithPass(email, password)
	if err != nil {
		if err.Code == http.StatusUnauthorized {
			attempts, err2 := s.getAttempts(email)
			attempts++
			if err2 := s.setAttempts(email, attempts); err2 != nil {
				err.Msg += fmt.Sprintf(";\n set attempts error: %v", err2.Trace())
				return nil, err.Trace()
			}
			if err2 != nil {
				err.Msg += fmt.Sprintf(";\n get attmpts error: %v", err2.Trace())
				return nil, err.Trace()
			}
			log.Println("attempts", attempts)
			if attempts > s.maxLoginAttempts {
				unblockTime = time.Now().Add(s.blockDuration)
				if err2 := s.blockUser(email); err2 != nil {
					err.Msg += fmt.Sprintf(";\n block user error: %v", err2.Trace())
					return nil, err.Trace()
				}
				return nil, errors.New(fmt.Sprintf("user %s blocked", email),
					fmt.Sprintf("exceeded number of login attempts, next attempt - %s", unblockTime.Format("01.02 15:04")),
					http.StatusTooManyRequests)
			}
		}
		return nil, err.Trace()
	}
	s.resetAttempts(email)

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
		return nil, errors.New1Msg("invalid refresh token", http.StatusUnauthorized)
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
		RefreshToken:          newRefresh,
		AccessTokenExpiresAt:  accessExpired,
		RefreshTokenExpiresAt: refreshExpired,
	}, err
}

func (s *AuthService) newRefreshToken() string {
	uid, _ := uuid.NewUUID()
	return uid.String()
}

func (s *AuthService) getAttempts(email string) (int, *errors.Error) {
	attempts, err := s.cache.Get("attempts-" + email)
	if err != nil {
		return 0, err.Trace()
	}
	return attempts, nil
}

func (s *AuthService) setAttempts(email string, attempts int) *errors.Error {
	if err := s.cache.Set("attempts-"+email, attempts, s.blockDuration); err != nil {
		return err.Trace()
	}
	return nil
}

func (s *AuthService) resetAttempts(email string) *errors.Error {
	if err := s.cache.Del("attempts-" + email); err != nil {
		return err.Trace()
	}
	return nil
}

func (s *AuthService) isBlocked(email string) (bool, time.Time, *errors.Error) {
	unblockTUnix, err := s.cache.Get("blocked-" + email)
	if err != nil {
		return false, time.Time{}, err.Trace()
	}
	if unblockTUnix == 0 {
		return false, time.Time{}, nil
	}
	unblockTime := time.Unix(int64(unblockTUnix), 0)
	return true, unblockTime, nil
}

func (s *AuthService) blockUser(email string) *errors.Error {
	if err := s.cache.Set("blocked-"+email, int(time.Now().Add(s.blockDuration).Unix()), s.blockDuration); err != nil {
		return err.Trace()
	}
	return nil
}
