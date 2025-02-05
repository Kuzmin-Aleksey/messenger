package auth

import (
	"context"
	"github.com/stretchr/testify/mock"
	"messanger/config"
	"messanger/domain/models"
	"messanger/domain/ports/mocks"
	mocks2 "messanger/domain/service/auth/mocks"
	"testing"
)

var authCfg = &config.AuthServiceConfig{
	AccessTokenTTLMin:    5,
	RefreshTokenTTLDays:  15,
	AccessTokenSignKey:   "test-key",
	DurationBlockUserMin: 5,
	LoginAttempts:        1,
}

func TestAuth(t *testing.T) {

	user := &models.User{
		Id:        1,
		Phone:     "70000000000",
		Password:  "123456",
		Name:      "name",
		RealName:  "Aleksey",
		ShowPhone: true,
		Confirmed: true,
	}

	phoneCode := "12345"

	cache := mocks.NewCache(t)
	userRepo := mocks.NewUsersRepo(t)
	phoneConf := mocks2.NewPhoneConfirmator(t)

	cache.On("Get", mock.Anything, user.Phone).Return(0, nil)
	cache.On("Del", context.Background(), user.Phone).Return(nil)

	userRepo.On("GetByPhoneWithPass", mock.Anything, user.Phone, user.Password).Return(user, nil)
	userRepo.On("GetById", mock.Anything, user.Id).Return(user, nil)

	phoneConf.On("ToConfirming", mock.Anything, user.Id, user.Phone).Return(nil)
	phoneConf.On("ConfirmUser", mock.Anything, phoneCode).Return(user.Id, nil)

	auth := NewAuthService(cache, userRepo, phoneConf, authCfg)

	if err := auth.Login1FA(context.Background(), user.Phone, user.Password); err != nil {
		t.Error(err)
	}

}
