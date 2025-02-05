package sms

import (
	"fmt"
	"messanger/pkg/errors"
)

type SmsChan struct {
	Chan chan string
}

func NewSmsChan() *SmsChan {
	return &SmsChan{
		Chan: make(chan string, 10),
	}
}

func (s *SmsChan) Send(to string, msg string) *errors.Error {
	s.Chan <- msg
	fmt.Println(to, msg)
	return nil
}
