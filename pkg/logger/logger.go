package logger

import (
	"io"
	"log"
	"sync"

	"github.com/cstone-io/twine/pkg/config"
	"github.com/cstone-io/twine/pkg/errors"
)

var (
	once     sync.Once
	instance *Logger
)

// Logger provides structured logging with multiple severity levels
type Logger struct {
	traceLogger    *log.Logger
	debugLogger    *log.Logger
	infoLogger     *log.Logger
	warnLogger     *log.Logger
	errorLogger    *log.Logger
	criticalLogger *log.Logger
	level          config.LogLevel
}

// Get returns the singleton logger instance
func Get() *Logger {
	once.Do(func() {
		cfg := config.Get()
		initialize(cfg.Logger)
	})
	return instance
}

func initialize(cfg config.LoggerConfig) {
	logfmt := log.Ldate | log.Ltime | log.Lshortfile
	instance = &Logger{
		traceLogger:    log.New(io.MultiWriter(cfg.Output), "TRACE: ", logfmt),
		debugLogger:    log.New(io.MultiWriter(cfg.Output), "DEBUG: ", logfmt),
		infoLogger:     log.New(io.MultiWriter(cfg.Output), "INFO: ", logfmt),
		warnLogger:     log.New(io.MultiWriter(cfg.Output), "WARN: ", logfmt),
		errorLogger:    log.New(io.MultiWriter(cfg.ErrorOutput), "ERROR: ", logfmt),
		criticalLogger: log.New(io.MultiWriter(cfg.ErrorOutput), "CRITICAL: ", logfmt),
		level:          cfg.Level,
	}
}

// Trace logs trace-level messages
func (l *Logger) Trace(format string, v ...interface{}) {
	if l.level <= config.LogTrace {
		l.traceLogger.Printf(format, v...)
	}
}

// Debug logs debug-level messages
func (l *Logger) Debug(format string, v ...interface{}) {
	if l.level <= config.LogDebug {
		l.debugLogger.Printf(format, v...)
	}
}

// Info logs info-level messages
func (l *Logger) Info(format string, v ...interface{}) {
	if l.level <= config.LogInfo {
		l.infoLogger.Printf(format, v...)
	}
}

// Warn logs warning-level messages
func (l *Logger) Warn(format string, v ...interface{}) {
	if l.level <= config.LogWarn {
		l.warnLogger.Printf(format, v...)
	}
}

// Error logs error-level messages
func (l *Logger) Error(format string, v ...interface{}) {
	if l.level <= config.LogError {
		l.errorLogger.Printf(format, v...)
	}
}

// Critical logs critical-level messages (always logged)
func (l *Logger) Critical(format string, v ...interface{}) {
	l.criticalLogger.Printf(format, v...)
}

// CustomError logs a structured error based on its severity
func (l *Logger) CustomError(e *errors.Error) {
	switch e.Severity {
	case errors.ErrMinor:
		l.Warn("%s", e.ErrorChain())
	case errors.ErrError:
		l.Error("%s", e.ErrorChain())
	case errors.ErrCritical:
		l.Critical("%s", e.ErrorChain())
	}
}
