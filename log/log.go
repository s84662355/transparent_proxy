package log

import (
	"runtime/debug"

	"go.uber.org/zap"
)

func Info(msg string, fields ...zap.Field) {
	if zaploger == nil {
		return
	}
	zaploger.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	if zaploger == nil {
		return
	}
	zaploger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	if zaploger == nil {
		return
	}
	zaploger.Error(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	if zaploger == nil {
		return
	}
	zaploger.Debug(msg, fields...)
}

func DPanic(msg string, fields ...zap.Field) {
	if zaploger == nil {
		return
	}
	zaploger.DPanic(msg, fields...)
}

func Panic(msg string, fields ...zap.Field) {
	if zaploger == nil {
		return
	}
	zaploger.Panic(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	if zaploger == nil {
		return
	}
	zaploger.Fatal(msg, fields...)
}

// /会抛出异常
func Recover(msg string, fields ...zap.Field) {
	if zaploger == nil {
		return
	}

	err := recover()
	if err != nil {
		zaploger.Fatal(msg, zap.Any("recover", err), zap.Any("debug.Stack", string(debug.Stack())))
	}
}

// /不会抛出异常
func DRecover(msg string, fields ...zap.Field) {
	if zaploger == nil {
		return
	}
	err := recover()
	if err != nil {
		zaploger.DPanic(msg, zap.Any("recover", err), zap.Any("debug.Stack", string(debug.Stack())))
	}
}
