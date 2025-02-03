package sms

import (
	"fmt"
	"messanger/pkg/errors"
)

type CmdSmsAdapter struct {
}

func NewCmdSmsAdapter() *CmdSmsAdapter {
	return &CmdSmsAdapter{}
}

func (s *CmdSmsAdapter) Send(to string, text string) *errors.Error {
	fmt.Printf("Sms Sender: to: %s, text: %s\n", to, text)
	return nil
}
