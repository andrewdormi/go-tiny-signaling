package signaling

import (
	"runtime"
	"runtime/debug"
)

func Recover(flag string) {
	_, _, _, _ = runtime.Caller(1)
	if err := recover(); err != nil {
		debug.PrintStack()
	}
}
