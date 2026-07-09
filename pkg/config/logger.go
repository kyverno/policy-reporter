package config

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// SetupLogger resolver method
func SetupLogger(c *Config) (*zap.Logger, error) {
	encoder := zap.NewProductionEncoderConfig()
	if c.Logging.Development {
		encoder = zap.NewDevelopmentEncoderConfig()
		encoder.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	}

	output := "json"
	if c.Logging.Encoding != "json" {
		output = "console"
		encoder.EncodeCaller = nil
	}

	var sampling *zap.SamplingConfig
	if !c.Logging.Development {
		sampling = &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		}
	}

	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(zapcore.Level(c.Logging.LogLevel)),
		Development:       c.Logging.Development,
		Sampling:          sampling,
		Encoding:          output,
		EncoderConfig:     encoder,
		DisableStacktrace: !c.Logging.Development,
		OutputPaths:       []string{"stderr"},
		ErrorOutputPaths:  []string{"stderr"},
	}

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	zap.ReplaceGlobals(logger)

	return logger, nil
}
