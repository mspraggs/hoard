package util

import (
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logConfig = func() zap.Config {
	config := zap.NewProductionConfig()
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig = encoderConfig
	return config
}()

// LogConfigOption defines the function signature used to modify the global
// logging configuration.
type LogConfigOption func(*zap.Config)

// MustNewLogger instantiates a new zap logger with package-specific logging
// options.
func MustNewLogger(options ...zap.Option) *zap.SugaredLogger {
	logger, err := logConfig.Build(options...)
	if err != nil {
		log.Panicf("Unable to instantiate logger: %v", err)
	}
	return logger.Sugar()
}

// ConfigureLogging applies the provided LogConfigOption to the internal logger
// configuration.
func ConfigureLogging(options ...LogConfigOption) {
	for _, option := range options {
		option(&logConfig)
	}
}

// WithLogLevel creates a logging configuration option that sets the specified
// logging level.
func WithLogLevel(level zapcore.Level) LogConfigOption {
	return func(c *zap.Config) {
		c.Level = zap.NewAtomicLevelAt(level)
	}
}

// WithOutputFilePath creates a logging configuration option that specifies an
// output logging file with the specified path.
func WithOutputFilePath(path string) LogConfigOption {
	return func(c *zap.Config) {
		c.OutputPaths = []string{path}
	}
}
