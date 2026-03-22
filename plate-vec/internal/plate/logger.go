package plate

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	reset     = "\x1b[0m"
	gray      = "\x1b[90m"
	blue      = "\x1b[34m"
	red       = "\x1b[31m"
	yellow    = "\x1b[33m"
	greenRGB  = "\x1b[0;38;2;104;173;0;49m"
	prefixVec = "[plate::vec]"
)

func timestampAndCaller() string {
	now := time.Now()
	timestamp := fmt.Sprintf(
		"%02d:%02d:%02d.%03d",
		now.Hour(),
		now.Minute(),
		now.Second(),
		now.Nanosecond()/1_000_000,
	)

	for skip := 2; skip < 12; skip++ {
		pc, file, line, ok := runtime.Caller(skip)
		if !ok {
			continue
		}

		fn := runtime.FuncForPC(pc)
		fnName := "unknown"
		if fn != nil {
			name := fn.Name()
			if idx := strings.LastIndex(name, "."); idx >= 0 && idx < len(name)-1 {
				fnName = name[idx+1:]
			} else {
				fnName = name
			}
		}

		base := filepath.Base(file)
		if base == "logger.go" || base == "log.go" {
			continue
		}

		caller := fmt.Sprintf("%s:%d %s", base, line, fnName)
		return gray + timestamp + " <" + greenRGB + caller + gray + ">" + reset
	}

	return "[" + timestamp + "]"
}

func printWithColor(color string, args ...interface{}) {
	parts := []interface{}{color + prefixVec + reset, timestampAndCaller()}
	parts = append(parts, args...)
	fmt.Println(parts...)
}

func Info(args ...interface{}) {
	printWithColor(blue, args...)
}

func Log(args ...interface{}) {
	Info(args...)
}

func Warn(args ...interface{}) {
	printWithColor(yellow, args...)
}

func Error(args ...interface{}) {
	printWithColor(red, args...)
}

func TryCatch[T any](fn func() T, errorMessage string) (result T, ok bool) {
	defer func() {
		if r := recover(); r != nil {
			Error(errorMessage, r)
			var zero T
			result = zero
			ok = false
		}
	}()

	return fn(), true
}

func TryCatchAsync[T any](fn func() (T, error), errorMessage string) (T, bool) {
	result, err := fn()
	if err != nil {
		Error(errorMessage, err)
		var zero T
		return zero, false
	}
	return result, true
}
