package sms

import "messanger/pkg/errors"

// жду одобрения заявки яндекс ...

type YandexClient struct {
}

func (c *YandexClient) Send(phone string, msg string) *errors.Error {
	return nil
}
