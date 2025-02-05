package auth

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"messanger/config"
	"messanger/domain/models"
	"messanger/domain/ports"
	"messanger/pkg/errors"
	"net/http"
	"time"
)

const ErrInvalidToken = "invalid token"

type AuthService struct {
	cache     ports.Cache
	repo      ports.UsersRepo
	phoneConf PhoneConfirmator

	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	accessTokenKey  []byte

	blockDuration    time.Duration
	maxLoginAttempts int
}

type PhoneConfirmator interface {
	ToConfirming(ctx context.Context, userId int, phone string) *errors.Error
	ConfirmUser(ctx context.Context, code string) (int, *errors.Error)
}

func NewAuthService(cache ports.Cache, repo ports.UsersRepo, phoneConf PhoneConfirmator, cfg *config.AuthServiceConfig) *AuthService {
	accessTokenTTL := time.Duration(cfg.AccessTokenTTLMin) * time.Minute
	refreshTokenTTL := time.Duration(cfg.RefreshTokenTTLDays) * time.Hour * 24
	blockDuration := time.Duration(cfg.DurationBlockUserMin) * time.Minute

	return &AuthService{
		cache:            cache,
		repo:             repo,
		phoneConf:        phoneConf,
		refreshTokenTTL:  refreshTokenTTL,
		accessTokenTTL:   accessTokenTTL,
		accessTokenKey:   []byte(cfg.AccessTokenSignKey),
		blockDuration:    blockDuration,
		maxLoginAttempts: cfg.LoginAttempts,
	}
}

func (s *AuthService) Login1FA(ctx context.Context, phone, password string) *errors.Error {
	attempts, err := s.cache.Get(ctx, phone)
	if err != nil {
		return err.Trace()
	}
	if attempts >= s.maxLoginAttempts {
		ttl := s.cache.TTL(ctx, phone)

		return errors.New(fmt.Sprintf("user %s blocked", phone),
			fmt.Sprintf("exceeded number of login attempts, next attempt is at - %s",
				time.Now().Add(ttl).Format("15:04:05")),
			http.StatusTooManyRequests)
	}

	user, err := s.repo.GetByPhoneWithPass(ctx, phone, password)
	if err != nil {
		if err.Code == http.StatusUnauthorized {
			attempts++
			if err2 := s.cache.Set(ctx, phone, attempts, s.blockDuration); err2 != nil {
				err.Msg += fmt.Sprintf("\n set attempts error: %v", err2.Trace())
			}
		}
		return err.Trace()
	}
	s.cache.Del(ctx, phone)

	if err := s.phoneConf.ToConfirming(ctx, user.Id, user.Phone); err != nil {
		return err.Trace()
	}

	return nil
}

func (s *AuthService) Login2FA(ctx context.Context, phone string, code string) (*models.Tokens, *errors.Error) {
	userId, err := s.phoneConf.ConfirmUser(ctx, code)
	if err != nil {
		return nil, err.Trace()
	}

	user, err := s.repo.GetById(ctx, userId)
	if err != nil {
		return nil, err.Trace()
	}

	if phone != user.Phone {
		return nil, errors.New1Msg("invalid phone number", http.StatusUnauthorized)
	}

	access, accessExpires, err := s.NewAccessToken(userId)
	if err != nil {
		return nil, err.Trace()
	}

	refreshExpired := time.Now().Add(s.refreshTokenTTL)
	refresh := s.newRefreshToken()
	if err := s.cache.Set(ctx, refresh, userId, s.refreshTokenTTL); err != nil {
		return nil, err.Trace()
	}

	return &models.Tokens{
		AccessToken:           access,
		RefreshToken:          refresh,
		AccessTokenExpiresAt:  accessExpires,
		RefreshTokenExpiresAt: refreshExpired,
	}, nil
}

func (s *AuthService) UpdateTokens(ctx context.Context, refresh string) (*models.Tokens, *errors.Error) {
	if len(refresh) == 0 {
		return nil, errors.New1Msg("missing refresh token", http.StatusBadRequest)
	}
	userId, err := s.cache.Get(ctx, refresh)
	if err != nil {
		return nil, err.Trace()
	}
	if userId == 0 {
		return nil, errors.New1Msg("invalid refresh token", http.StatusUnauthorized)
	}

	if err := s.cache.Del(ctx, refresh); err != nil {
		return nil, err.Trace()
	}

	access, accessExpired, err := s.NewAccessToken(userId)
	if err != nil {
		return nil, err.Trace()
	}

	refreshExpired := time.Now().Add(s.refreshTokenTTL)
	newRefresh := s.newRefreshToken()
	if err := s.cache.Set(ctx, newRefresh, userId, s.refreshTokenTTL); err != nil {
		return nil, err.Trace()
	}

	return &models.Tokens{
		AccessToken:           access,
		RefreshToken:          newRefresh,
		AccessTokenExpiresAt:  accessExpired,
		RefreshTokenExpiresAt: refreshExpired,
	}, err
}

func (s *AuthService) DecodeAccessToken(access string) (int, *errors.Error) {
	claims, err := jwt.Parse(access, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return s.accessTokenKey, nil
	})
	_ = "transaction"
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

func (s *AuthService) NewAccessToken(id int) (string, time.Time, *errors.Error) {
	expires := time.Now().Add(s.accessTokenTTL)

	claims := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"id":      id,
		"expires": expires.Unix(),
	})
	token, err := claims.SignedString(s.accessTokenKey)
	if err != nil {
		return "", expires, errors.New(err, "create token error", http.StatusBadRequest)
	}
	return token, expires, nil
}

func (s *AuthService) newRefreshToken() string {
	uid, _ := uuid.NewUUID()
	return uid.String()
}
