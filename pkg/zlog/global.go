package zlog

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// GetLevel gets the global logging level.
func GetLevel() zapcore.Level {
	return gP.Load().(*ZapProperties).Level.Level()
}

// SetLevel alters the global logging level.
func SetLevel(l zapcore.Level) {
	gP.Load().(*ZapProperties).Level.SetLevel(l)
}

// With creates a child logger and adds structured context to it.
// Fields added to the child don't affect the parent, and vice versa.
func With(fields ...zap.Field) *zap.Logger {
	return L().WithOptions(zap.AddCallerSkip(1)).With(fields...)
}

// Debug logs a message at DebugLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Debug(msg string, fields ...zap.Field) {
	L().WithOptions(zap.AddCallerSkip(1)).Debug(msg, fields...)
}

// Info logs a message at InfoLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Info(msg string, fields ...zap.Field) {
	L().WithOptions(zap.AddCallerSkip(1)).Info(msg, fields...)
}

// Warn logs a message at WarnLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Warn(msg string, fields ...zap.Field) {
	L().WithOptions(zap.AddCallerSkip(1)).Warn(msg, fields...)
}

// Error logs a message at ErrorLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Error(msg string, fields ...zap.Field) {
	L().WithOptions(zap.AddCallerSkip(1)).Error(msg, fields...)
}

// Panic logs a message at PanicLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then panics, even if logging at PanicLevel is disabled.
func Panic(msg string, fields ...zap.Field) {
	L().WithOptions(zap.AddCallerSkip(1)).Panic(msg, fields...)
}

// Fatal logs a message at FatalLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then calls os.Exit(1), even if logging at FatalLevel is
// disabled.
func Fatal(msg string, fields ...zap.Field) {
	L().WithOptions(zap.AddCallerSkip(1)).Fatal(msg, fields...)
}

func Debugf(format string, args ...interface{}) {
	L().WithOptions(zap.AddCallerSkip(1)).Sugar().Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	L().WithOptions(zap.AddCallerSkip(1)).Sugar().Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	L().WithOptions(zap.AddCallerSkip(1)).Sugar().Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	L().WithOptions(zap.AddCallerSkip(1)).Sugar().Errorf(format, args...)
}

func DPanicf(format string, args ...interface{}) {
	L().WithOptions(zap.AddCallerSkip(1)).Sugar().DPanicf(format, args...)
}

func Panicf(format string, args ...interface{}) {
	L().WithOptions(zap.AddCallerSkip(1)).Sugar().Panicf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	L().WithOptions(zap.AddCallerSkip(1)).Sugar().Fatalf(format, args...)
}
