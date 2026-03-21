package main

import "plate/db/internal/plate"

func Info(args ...interface{}) {
	plate.Info(args...)
}

func Log(args ...interface{}) {
	plate.Log(args...)
}

func Warn(args ...interface{}) {
	plate.Warn(args...)
}

func Error(args ...interface{}) {
	plate.Error(args...)
}

func TryCatch[T any](fn func() T, errorMessage string) (result T, ok bool) {
	return plate.TryCatch(fn, errorMessage)
}

func TryCatchAsync[T any](fn func() (T, error), errorMessage string) (T, bool) {
	return plate.TryCatchAsync(fn, errorMessage)
}
