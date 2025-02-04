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
		TracePath:   getTrace(2),
		Msg:         fmt.Sprint(msg),
		UserMessage: fmt.Sprint(userMsg),
		Code:        code,
	}
}

func New1Msg(msg any, code int) *Error {
	return &Error{
		TracePath:   getTrace(2),
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
	e.TracePath = getTrace(2) + " > " + e.TracePath
	return e
}

func Trace(err error) error {
	return errors.New(getTrace(2) + " > " + err.Error())
}

func getTrace(skip int) string {
	pc, f, l, _ := runtime.Caller(skip)
	callerFunc := runtime.FuncForPC(pc)
	callerFunc.Name()
	return trimPath(f) + ":" + strconv.Itoa(l) + " " + trimFuncPath(callerFunc.Name())
}

func trimPath(path string) string {
	s := false

	for i := len(path) - 1; i != 0; i-- {
		if path[i] == '/' {
			if s {
				return path[i:]
			}
			s = true
		}
	}
	return path
}

func trimFuncPath(path string) string {
	for i := len(path) - 1; i != 0; i-- {
		if path[i] == '.' {
			return path[i:]
		}
	}
	return path
}
