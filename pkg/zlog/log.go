package zlog

import (
	"errors"
	"fmt"
	"github.com/jsternberg/zap-logfmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sync/atomic"
)

var gL, gP, gS atomic.Value

func init() {
	conf := &Config{Level: "info"}
	logger, props, _ := NewLogger(conf)
	ReplaceGlobals(logger, props)
}

// NewLogger initializes a zap logger.
func NewLogger(cfg *Config, opts ...zap.Option) (*zap.Logger, *ZapProperties, error) {
	var output zapcore.WriteSyncer
	if len(cfg.File.Filename) > 0 {
		return nil, nil, errors.New("unimplemented")
	} else {
		stderr, _, err := zap.Open("stderr")
		if err != nil {
			return nil, nil, err
		}
		output = stderr
	}
	return NewLoggerWithSyncer(cfg, output, opts...)
}

// NewLoggerWithSyncer initializes a zap logger with specified write syncer.
func NewLoggerWithSyncer(cfg *Config, output zapcore.WriteSyncer, opts ...zap.Option) (*zap.Logger, *ZapProperties, error) {
	level := zap.NewAtomicLevel()
	err := level.UnmarshalText([]byte(cfg.Level))
	if err != nil {
		return nil, nil, err
	}
	encoder, err := newEncoder(cfg)
	if err != nil {
		return nil, nil, err
	}
	core := zapcore.NewCore(encoder, output, level)
	opts = append(cfg.buildOptions(output), opts...)
	lg := zap.New(core, opts...)
	prop := &ZapProperties{
		Core:   core,
		Syncer: output,
		Level:  level,
	}
	return lg, prop, nil
}

func newEncoder(cfg *Config) (zapcore.Encoder, error) {
	encConfig := zap.NewProductionEncoderConfig()
	if cfg.Development {
		encConfig = zap.NewDevelopmentEncoderConfig()
	}
	switch cfg.Format {
	case "json":
		return zapcore.NewJSONEncoder(encConfig), nil
	case "console":
		return zapcore.NewConsoleEncoder(encConfig), nil
	case "logfmt", "":
		return zaplogfmt.NewEncoder(encConfig), nil
	default:
		return nil, fmt.Errorf("unknown logging format %s", cfg.Format)
	}
}

// L returns the global Logger, which can be reconfigured with
// ReplaceGlobals. It's safe for concurrent use.
func L() *zap.Logger {
	return gL.Load().(*zap.Logger)
}

// S returns the global SugaredLogger, which can be reconfigured with
// ReplaceGlobals. It's safe for concurrent use.
func S() *zap.SugaredLogger {
	return gS.Load().(*zap.SugaredLogger)
}

// ReplaceGlobals replaces the global Logger and SugaredLogger.
// It's safe for concurrent use.
func ReplaceGlobals(logger *zap.Logger, props *ZapProperties) {
	gL.Store(logger)
	gS.Store(logger.Sugar())
	gP.Store(props)
}

// Sync flushes any buffered log entries.
func Sync() error {
	err := L().Sync()
	if err != nil {
		return err
	}
	return S().Sync()
}
