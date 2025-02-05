package phone

import (
	"context"
	"fmt"
	"golang.org/x/exp/rand"
	"messanger/domain/ports"
	"messanger/pkg/errors"
	"net/http"
	"time"
)

type SmsSender interface {
	Send(phone, content string) *errors.Error
}

type PhoneService struct {
	cache   ports.Cache
	sms     SmsSender
	codeTTL time.Duration
}

func NewPhoneService(sms SmsSender, cache ports.Cache) *PhoneService {
	return &PhoneService{
		cache:   cache,
		sms:     sms,
		codeTTL: time.Hour,
	}
}

func (s *PhoneService) ToConfirming(ctx context.Context, userId int, phone string) *errors.Error {
	code, err := s.generateCode(ctx)
	if err != nil {
		return err.Trace()
	}
	if err := s.cache.Set(ctx, code, userId, s.codeTTL); err != nil {
		return err.Trace()
	}
	if err := s.sms.Send(phone, code); err != nil {
		return err.Trace()
	}
	return nil
}

func (s *PhoneService) ConfirmUser(ctx context.Context, code string) (int, *errors.Error) {
	userId, err := s.cache.Get(ctx, code)
	if err != nil {
		return 0, err.Trace()
	}
	if userId == 0 {
		return 0, errors.New1Msg("invalid code", http.StatusUnauthorized)
	}
	if err := s.cache.Del(ctx, code); err != nil {
		return 0, err.Trace()
	}
	return userId, nil
}

func (s *PhoneService) generateCode(ctx context.Context) (string, *errors.Error) {
	var code string
	for {
		codeN := rand.Intn(10000) // 4 digit in code
		code = fmt.Sprintf("%4d", codeN)

		v, err := s.cache.Get(ctx, code)
		if err != nil {
			return code, err.Trace()
		}
		if v == 0 {
			break
		}
	}
	return code, nil
}
