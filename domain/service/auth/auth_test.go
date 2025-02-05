package auth

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/mock"
	"messanger/config"
	cache "messanger/data/cache/local"
	sms "messanger/data/sms/sms_chan"
	"messanger/domain/models"
	"messanger/domain/ports"
	"messanger/domain/ports/mocks"
	"messanger/domain/service/phone"
	"testing"
)

var authCfg = &config.AuthServiceConfig{
	AccessTokenTTLMin:    5,
	RefreshTokenTTLDays:  15,
	AccessTokenSignKey:   "test-key",
	DurationBlockUserMin: 5,
	LoginAttempts:        1,
}

var c ports.Cache
var smsChan *sms.SmsChan
var phoneConf PhoneConfirmator

func init() {
	c = cache.NewCache()
	smsChan = sms.NewSmsChan()
	phoneConf = phone.NewPhoneService(smsChan, c)
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

	userRepo := mocks.NewUsersRepo(t)

	userRepo.On("GetByPhoneWithPass", mock.Anything, user.Phone, user.Password).Return(user, nil)
	userRepo.On("GetById", mock.Anything, user.Id).Return(user, nil)

	auth := NewAuthService(c, userRepo, phoneConf, authCfg)

	if err := auth.Login1FA(context.Background(), user.Phone, user.Password); err != nil {
		t.Error(err)
	}

	code := <-smsChan.Chan
	fmt.Println(user.Phone, code)
	tokens, err := auth.Login2FA(context.Background(), user.Phone, code)
	if err != nil {
		t.Error(err)
	}

	t.Log(tokens)

}
