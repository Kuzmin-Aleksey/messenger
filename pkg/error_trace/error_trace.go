package tr

import (
	"fmt"
	"runtime"
)

func Trace(err any) error {
	_, f, l, _ := runtime.Caller(1)
	return fmt.Errorf("%s:%d > %v", trimPath(f), l, err)
}

func trimPath(path string) string {
	const slash uint8 = 47
	s := false

	for i := len(path) - 1; i != 0; i-- {
		if path[i] == slash {
			if s {
				return path[i:]
			}
			s = true
		}
	}
	return path
}
