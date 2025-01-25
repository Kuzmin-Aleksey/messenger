package service

import (
	"crypto/tls"
	"fmt"
	"gopkg.in/gomail.v2"
	"messanger/config"
	"messanger/core/service/jwt"
	"messanger/pkg/errors"
	"net/http"
	"time"
)

type EmailService struct {
	dialer          *gomail.Dialer
	myEmail         string
	confirmEmailURL string
	tokenManager    *jwt.TokenManager
	OnConfirm       func(userId int) *errors.Error
}

func NewEmailService(cfg *config.EmailServiceConfig) *EmailService {
	dialer := gomail.NewDialer(cfg.SMTP.Server, cfg.SMTP.Port, cfg.SMTP.Email, cfg.SMTP.Password)
	dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	tokenManager := jwt.NewTokenManager(time.Duration(cfg.TTLMin)*time.Minute, []byte(cfg.TokenSignKey))
	return &EmailService{
		dialer:          dialer,
		tokenManager:    tokenManager,
		confirmEmailURL: cfg.ConfirmEmailURL,
		myEmail:         cfg.SMTP.Email,
	}
}

func (s *EmailService) SetOnConfirm(f func(userId int) *errors.Error) {
	s.OnConfirm = f
}

func (s *EmailService) ToConfirming(userId int, email string) *errors.Error {
	token, _, err := s.tokenManager.NewToken(userId)
	if err != nil {
		return err.Trace()
	}
	if err := s.sendConfirmEmail(token, email); err != nil {
		return err.Trace()
	}
	return nil
}

func (s *EmailService) ConfirmUser(token string) *errors.Error {
	if len(token) == 0 {
		return errors.New1Msg("missing token in form", http.StatusBadRequest)
	}
	if s.tokenManager == nil {
		return errors.New1Msg("tokenManager is nil", http.StatusBadRequest)
	}
	userId, err := s.tokenManager.DecodeToken(token)
	if err != nil {
		return err.Trace()
	}
	if err := s.OnConfirm(userId); err != nil {
		return err.Trace()
	}
	return nil
}

// todo - using template file
const template = `
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Подтвердите email</title>
    <style>
        body { font-family: Arial, sans-serif; }
        h1 { color: #333; }
        a { color: #007BFF; text-decoration: none; }
    </style>
</head>
<body>
    <h1>Здравствуйте!</h1>
    <p>Благодарим вас за регистрацию. Для подтверждения вашего email, перейдите по следующей ссылке:</p>
    <p><a href="%s?token=%s">Подтвердить email</a></p>
</body>
</html>
`

func (s *EmailService) sendConfirmEmail(token string, email string) *errors.Error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.myEmail)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Confirm email")
	m.SetBody("text/html", fmt.Sprintf(template, s.confirmEmailURL, token))

	if err := s.dialer.DialAndSend(m); err != nil {
		return errors.New(err, "send confirm email fail", http.StatusInternalServerError)
	}
	return nil
}
