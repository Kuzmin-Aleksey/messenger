package smsarea

import (
	"encoding/json"
	errorsutils "errors"
	"fmt"
	"io"
	"messanger/config"
	"messanger/pkg/errors"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// 33UTXSb6TvspeJDRoXGFEdYThEH4GNAZ

type SmsSender struct {
	login  string
	key    string
	sign   string
	client *http.Client
}

func NewSmsSender(cfg *config.SmsSenderConfig) (*SmsSender, error) {
	s := &SmsSender{
		login: cfg.Login,
		key:   cfg.Key,
		sign:  cfg.Sign,
		client: &http.Client{
			Timeout: time.Duration(cfg.TimeoutSec) * time.Second,
		},
	}
	if err := s.TestAuth(); err != nil {
		return nil, errors.Trace(err)
	}
	return s, nil
}

type authResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (s *SmsSender) TestAuth() error {
	req, err := http.NewRequest("GET", s.getAuthUrl(), nil)
	if err != nil {
		return errors.Trace(errorsutils.New("sms auth: make request error: " + err.Error()))
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return errors.Trace(errorsutils.New("sms auth: http error: " + err.Error()))
	}
	status := new(authResponse)
	if err := json.NewDecoder(resp.Body).Decode(status); err != nil {
		return errors.Trace(errorsutils.New("sms auth: invalid response: " + err.Error()))
	}
	if !status.Success {
		return errors.Trace(errorsutils.New("sms auth: auth failed: " + status.Message))
	}
	return nil
}

func (s *SmsSender) getAuthUrl() string {
	return fmt.Sprintf("https://%s:%s@gate.smsaero.ru/v2/auth", s.login, s.key)
}

func (s *SmsSender) Send(phone, content string) *errors.Error {
	req, err := http.NewRequest("GET", s.getSendSmsUrl(phone, content), nil)
	if err != nil {
		return errors.New(err, "send sms failed", http.StatusInternalServerError)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return errors.New(err, "send sms failed", http.StatusInternalServerError)
	}
	fmt.Println(req.URL)
	resp.Write(os.Stdout)
	io.Copy(os.Stdout, resp.Body)
	return nil
}

func (s *SmsSender) getSendSmsUrl(phone, content string) string {
	return fmt.Sprintf("https://%s:%s@gate.smsaero.ru/v2/sms/send?number=%s&text=%s&sign=%s",
		url.QueryEscape(s.login),
		url.QueryEscape(s.key),
		url.QueryEscape(phone),
		url.QueryEscape(strings.ReplaceAll(content, " ", "+")),
		url.QueryEscape(s.sign))
}
