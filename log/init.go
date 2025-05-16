package log

import (
	"io"
	"os"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var encoder zapcore.Encoder

var infoLevel, warnLevel,
	errorLevel,
	debugLevel,
	dPanicLevel,
	panicLevel,
	fatalLevel zap.LevelEnablerFunc

var zaploger *zap.Logger

func Init(filepath string) {
	// 设置一些基本日志格式 具体含义还比较好理解，直接看zap源码也不难懂

	encoder = zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey:  "msg",
		LevelKey:    "level",
		EncodeLevel: zapcore.CapitalLevelEncoder,
		CallerKey:   "file",
		TimeKey:     "ts",
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05"))
		},
		SkipLineEnding: false,
		LineEnding:     "",
		FunctionKey:    "func",
		EncodeCaller:   zapcore.FullCallerEncoder,
		EncodeDuration: func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendInt64(int64(d) / 1000000)
		},
	})

	// 获取 info、warn日志文件的io.Writer 抽象 getWriter() 在下方实现
	infoWriter := getWriter(filepath + "/info")
	warnWriter := getWriter(filepath + "/warn")
	errorWriter := getWriter(filepath + "/error")
	debugWriter := getWriter(filepath + "/debug")
	dPanicWriter := getWriter(filepath + "/dPanic")
	panicWriter := getWriter(filepath + "/panic")
	fatalWriter := getWriter(filepath + "/fatal")

	infoLevel = zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zapcore.InfoLevel
	})
	warnLevel = zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zapcore.WarnLevel
	})
	errorLevel = zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zapcore.ErrorLevel
	})

	debugLevel = zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zapcore.DebugLevel
	})

	dPanicLevel = zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zapcore.DPanicLevel
	})

	panicLevel = zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zapcore.PanicLevel
	})

	fatalLevel = zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zapcore.FatalLevel
	})

	// 最后创建具体的Logger
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, zapcore.AddSync(infoWriter), infoLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(warnWriter), warnLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(errorWriter), errorLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(debugWriter), debugLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(dPanicWriter), dPanicLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(panicWriter), panicLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(fatalWriter), fatalLevel),

		zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), infoLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), warnLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), errorLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), debugLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), dPanicLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), panicLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), fatalLevel),
	)

	zaploger = zap.New(
		core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	)
}

func getWriter(filename string) io.Writer {
	// 生成rotatelogs的Logger 实际生成的文件名 demo.log.YYmmddHH
	// demo.log是指向最新日志的链接
	// 保存7天内的日志，每1小时(整点)分割一次日志

	hook, err := rotatelogs.New(
		filename+"_%Y-%m-%d.log", // 没有使用go风格反人类的format格式
		rotatelogs.WithLinkName(filename),
		rotatelogs.WithMaxAge(time.Hour*24*3),
		rotatelogs.WithRotationTime(time.Hour),
	)
	if err != nil {
		panic(err)
	}

	return hook
}
