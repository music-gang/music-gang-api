package log

import (
	"context"
	"io"

	"github.com/music-gang/music-gang-api/app/service"
)

// Force implementation of LogService interface for Logger.
var _ service.LogService = (*Logger)(nil)

// Logger is a wrapper for multiple loggin system.
type Logger struct {
	// backends is the list of loggers that implement the LogService interface.
	backends []service.LogService
}

// AddBackend adds a logger to the list of loggers.
func (l *Logger) AddBackend(logger service.LogService) {
	l.backends = append(l.backends, logger)
}

// Format returns the format of the log message.
func (l *Logger) Format() string {
	return service.FormatStandard
}

// Level returns the current standard log level.
// This wrapper func returns the max level of all loggers.
func (l *Logger) Level() int {
	maxLevel := service.LevelAll
	for _, logger := range l.backends {
		if logger.Level() > maxLevel {
			maxLevel = logger.Level()
		}
	}
	return maxLevel
}

// Output returns the writer used for standard output.
// This wrapper func returns os.Stdout.
func (l *Logger) Output() io.Writer {
	return io.Discard
}

// ReportDebug logs a message at level Debug.
// This wrapper func calls ReportDebug on all loggers that are at least at level Debug.
func (l *Logger) ReportDebug(ctx context.Context, msg string) {
	for _, logger := range l.backends {
		if logger.Level() <= service.LevelDebug {
			logger.ReportDebug(ctx, msg)
		}
	}
}

// ReportError logs an error.
// This wrapper func calls ReportError on all loggers that are at least at level Error.
func (l *Logger) ReportError(ctx context.Context, err error) {
	for _, logger := range l.backends {
		if logger.Level() <= service.LevelError {
			logger.ReportError(ctx, err)
		}
	}
}

// ReportFatal logs a fatal error.
// This wrapper func calls ReportFatal on all loggers that are at least at level Fatal.
func (l *Logger) ReportFatal(ctx context.Context, err error) {
	for _, logger := range l.backends {
		if logger.Level() <= service.LevelFatal {
			logger.ReportFatal(ctx, err)
		}
	}
}

// ReportInfo logs an info.
// This wrapper func calls ReportInfo on all loggers that are at least at level Info.
func (l *Logger) ReportInfo(ctx context.Context, info string) {
	for _, logger := range l.backends {
		if logger.Level() <= service.LevelInfo {
			logger.ReportInfo(ctx, info)
		}
	}
}

// ReportPanic logs a panic.
// This wrapper func calls ReportPanic on all loggers that are at least at level Panic.
func (l *Logger) ReportPanic(ctx context.Context, err interface{}) {
	for _, logger := range l.backends {
		if logger.Level() <= service.LevelPanic {
			logger.ReportPanic(ctx, err)
		}
	}
}

// ReportWarning logs a warning.
// This wrapper func calls ReportWarning on all loggers that are at least at level Warn.
func (l *Logger) ReportWarning(ctx context.Context, warning string) {
	for _, logger := range l.backends {
		if logger.Level() <= service.LevelWarn {
			logger.ReportWarning(ctx, warning)
		}
	}
}
