package log

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/music-gang/music-gang-api/app"
	"github.com/music-gang/music-gang-api/app/service"
)

var _ service.LogService = (*StdOutputLogger)(nil)

// StdOutputLogger is a LogService that writes to standard output.
type StdOutputLogger struct {
	out io.Writer
	Lvl int
	Fmt string
}

// StdOutputLoggerOptions represents the options for a StdOutputLogger.
type StdOutputLoggerOptions struct {
	Level  int
	Format string
}

// NewStdOutputLogger returns a new StdOutputLogger.
// The default log level is LevelAll.
func NewStdOutputLogger() *StdOutputLogger {
	return &StdOutputLogger{
		Lvl: service.LevelAll,
		Fmt: service.FormatStandard,
		out: os.Stdout,
	}
}

// NewStdOutputLoggerWithConfig returns a new StdOutputLogger with the given config.
func NewStdOutputLoggerWithConfig(options StdOutputLoggerOptions) *StdOutputLogger {
	level := service.LevelAll
	if options.Level >= service.LevelAll && options.Level <= service.LevelOff {
		level = options.Level
	}

	format := service.FormatStandard
	if options.Format != "" {
		format = options.Format
	}

	return &StdOutputLogger{
		Lvl: level,
		Fmt: format,
	}
}

// Format returns the current format.
func (s *StdOutputLogger) Format() string {
	return s.Fmt
}

// Level returns the current standard log level.
func (s *StdOutputLogger) Level() int {
	return s.Lvl
}

// Output returns the writer used for standard output.
// In this case, it is os.Stdout.
func (s *StdOutputLogger) Output() io.Writer {
	return s.out
}

// ReportDebug logs a message at level Debug.
func (s *StdOutputLogger) ReportDebug(ctx context.Context, msg string) {
	if s.Level() <= service.LevelDebug {
		fmt.Fprintln(s.out, service.FormatOutputForReportFunc(service.LevelDebug, msg, app.TagsFromContext(ctx), s.Fmt))
	}
}

// ReportError logs an error.
func (s *StdOutputLogger) ReportError(ctx context.Context, err error) {
	select {
	case <-ctx.Done():
		s.ReportDebug(ctx, err.Error())
		return
	default:
	}
	if s.Level() <= service.LevelError {
		fmt.Fprintln(s.out, service.FormatOutputForReportFunc(service.LevelError, err.Error(), app.TagsFromContext(ctx), s.Fmt))
	}
}

// ReportFatal logs a fatal error.
func (s *StdOutputLogger) ReportFatal(ctx context.Context, err error) {
	if s.Level() <= service.LevelFatal {
		fmt.Fprintln(s.out, service.FormatOutputForReportFunc(service.LevelFatal, err.Error(), app.TagsFromContext(ctx), s.Fmt))
	}
}

// ReportInfo logs an info.
func (s *StdOutputLogger) ReportInfo(ctx context.Context, info string) {
	if s.Level() <= service.LevelInfo {
		fmt.Fprintln(s.out, service.FormatOutputForReportFunc(service.LevelInfo, info, app.TagsFromContext(ctx), s.Fmt))
	}
}

// ReportPanic logs a panic.
func (s *StdOutputLogger) ReportPanic(ctx context.Context, err interface{}) {
	if s.Level() <= service.LevelPanic {
		fmt.Fprintln(s.out, service.FormatOutputForReportFunc(service.LevelPanic, fmt.Sprintf("%v", err), app.TagsFromContext(ctx), s.Fmt))
	}
}

// ReportWarning logs a warning.
func (s *StdOutputLogger) ReportWarning(ctx context.Context, warning string) {
	if s.Level() <= service.LevelWarn {
		fmt.Fprintln(s.out, service.FormatOutputForReportFunc(service.LevelWarn, warning, app.TagsFromContext(ctx), s.Fmt))
	}
}
