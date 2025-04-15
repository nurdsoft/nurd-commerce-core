// Package log is based on uber zap
package log

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New returns a new logger
func New(opts ...Option) (*zap.SugaredLogger, error) {
	// default Options
	options := options{} //nolint:govet

	for _, o := range opts {
		o.apply(&options)
	}

	zapConfig := zap.NewProductionConfig()
	zapConfig.DisableCaller = true
	zapConfig.DisableStacktrace = true
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	zapLogger, err := zapConfig.Build()
	if err != nil {
		return nil, err
	}

	if options.fileLogEnabled {
		config := zap.NewProductionEncoderConfig()
		config.EncodeTime = zapcore.ISO8601TimeEncoder
		fileEncoder := zapcore.NewJSONEncoder(config)

		// TODO log file name can be taken from component name
		logFile, _ := os.OpenFile(filepath.Join("logs", "api.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		writer := zapcore.AddSync(logFile)
		defaultLogLevel := zapcore.DebugLevel
		core := zapcore.NewTee(
			zapcore.NewCore(fileEncoder, writer, defaultLogLevel),
		)
		//return zap.New(zapcore.NewTee(zapLogger.Core(), sentryCore)).Sugar(), nil

		return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)).Sugar(), nil
	}

	return zapLogger.Sugar(), nil
}
