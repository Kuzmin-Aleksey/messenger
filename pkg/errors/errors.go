package errors

import (
	"errors"
	"fmt"
	"runtime"
	"strconv"
)

type Error struct {
	TracePath   string
	Msg         string
	UserMessage string
	Code        int
}

func New(msg any, userMsg any, code int) *Error {
	return &Error{
		TracePath:   getCaller(),
		Msg:         fmt.Sprint(msg),
		UserMessage: fmt.Sprint(userMsg),
		Code:        code,
	}
}

func New1Msg(msg any, code int) *Error {
	return &Error{
		TracePath:   getCaller(),
		UserMessage: fmt.Sprint(msg),
		Code:        code,
	}
}

func (e *Error) Error() string {
	if len(e.Msg) == 0 {
		return e.TracePath + ":\n" + e.UserMessage
	}
	return e.TracePath + ":\n" + e.UserMessage + ": " + e.Msg
}

func (e *Error) Trace() *Error {
	e.TracePath = getCaller() + " > " + e.TracePath
	return e
}

func Trace(err error) error {
	return errors.New(getCaller() + " > " + err.Error())
}

func getCaller() string {
	pc, _, line, _ := runtime.Caller(2)
	return runtime.FuncForPC(pc).Name() + ":" + strconv.Itoa(line)
}
