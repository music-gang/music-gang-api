package service

import (
	"context"
	"fmt"
	"io"
	"runtime"
	"strings"
	"time"
)

const (
	// Log levels to control the logging output.
	LevelAll = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
	LevelPanic
	LevelOff

	// codes for log levels.
	CodeDebug = "DEBUG"
	CodeInfo  = "INFO"
	CodeWarn  = "WARN"
	CodeError = "ERROR"
	CodeFatal = "FATAL"
	CodePanic = "PANIC"

	FormatStandard = "[%s][%s][%s] %s func=%s file=%s:%d"                                                                          // [2006-01-02 13:14:15][DEBUG][HTTP] hello world func=main.main file=main.go:12
	FormatMinimal  = "[%s][%s][%s] %s"                                                                                             // [2006-01-02 13:14:15][DEBUG][HTTP] hello world
	FormatJSON     = "{\"time\":\"%s\",\"level\":\"%s\",\"context\":\"%s\",\"message\":\"%s\",\"func\":\"%s\",\"file\":\"%s:%d\"}" // {"time":"2006-01-02T15:04:05Z07:00","level":"DEBUG","context":"HTTP","message":"hello world","func":"main.main","file":"main.go:12"}
)

var (
	// map of log levels to codes.
	codes = map[int]string{
		LevelDebug: CodeDebug,
		LevelInfo:  CodeInfo,
		LevelWarn:  CodeWarn,
		LevelError: CodeError,
		LevelFatal: CodeFatal,
		LevelPanic: CodePanic,
	}
)

// LogService describes an interface for logging.
type LogService interface {
	// Format returns the format of the log message.
	Format() string

	// Level returns the current standard log level.
	Level() int

	// Output returns the writer used for standard output.
	Output() io.Writer

	// ReportDebug logs a message at level Debug.
	ReportDebug(ctx context.Context, msg string)

	// ReportError logs an error.
	ReportError(ctx context.Context, err error)

	// ReportFatal logs a fatal error.
	ReportFatal(ctx context.Context, err error)

	// ReportInfo logs an info.
	ReportInfo(ctx context.Context, info string)

	// ReportPanic logs a panic.
	ReportPanic(ctx context.Context, err interface{})

	// ReportWarning logs a warning.
	ReportWarning(ctx context.Context, warning string)
}

// LogCode returns the code for the given log level.
func LogCode(level int) string {
	if code, ok := codes[level]; ok {
		return code
	}
	return ""
}

// LogLevel returns the log level for the given code.
func LogLevel(code string) int {
	for level, c := range codes {
		if code == c {
			return level
		}
	}
	return LevelAll
}

// FormatOutputForReportFunc formats the output for the given log level and message.
// Accepted types for toReport: string, error, fmt.Stringer, nil.
// format is the format of the log message, use one of the Format* constants inside this package.
func FormatOutputForReportFunc(level int, toReport interface{}, tags []string, format string) string {

	var output string

	pc, _, _, ok := runtime.Caller(3)
	if !ok {
		return ""
	}

	details := runtime.FuncForPC(pc)
	filename, line := details.FileLine(pc)

	switch data := toReport.(type) {
	case string:
		output = data
	case error:
		output = data.Error()
	case fmt.Stringer:
		output = data.String()
	case nil:
		output = "nil"
	default:
		return ""
	}

	implodedTags := ""
	if len(tags) > 0 {
		implodedTags = strings.Join(tags, ", ")
	}

	return fmt.Sprintf(format, time.Now().UTC().Format("2006-01-02 15:04:05"), LogCode(level), implodedTags, output, details.Name(), filename, line)
}
