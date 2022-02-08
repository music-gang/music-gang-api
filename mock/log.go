package mock

import (
	"context"
	"io"

	"github.com/music-gang/music-gang-api/app/service"
)

var _ service.LogService = (*LogService)(nil)

type LogService struct {
	FormatFn        func() string
	LevelFn         func() int
	OutputFn        func() io.Writer
	ReportDebugFn   func(ctx context.Context, msg string)
	ReportErrorFn   func(ctx context.Context, err error)
	ReportFatalFn   func(ctx context.Context, err error)
	ReportInfoFn    func(ctx context.Context, info string)
	ReportPanicFn   func(ctx context.Context, err interface{})
	ReportWarningFn func(ctx context.Context, warning string)
}

func (l *LogService) Format() string {
	if l.FormatFn == nil {
		panic("FormatFn not defined")
	}
	return l.FormatFn()
}

func (l *LogService) Level() int {
	if l.LevelFn == nil {
		panic("LevelFn not defined")
	}
	return l.LevelFn()
}

func (l *LogService) Output() io.Writer {
	if l.OutputFn == nil {
		panic("OutputFn not defined")
	}
	return l.OutputFn()
}

func (l *LogService) ReportDebug(ctx context.Context, msg string) {
	if l.ReportDebugFn == nil {
		panic("ReportDebugFn not defined")
	}
	l.ReportDebugFn(ctx, msg)
}

func (l *LogService) ReportError(ctx context.Context, err error) {
	if l.ReportErrorFn == nil {
		panic("ReportErrorFn not defined")
	}
	l.ReportErrorFn(ctx, err)
}

func (l *LogService) ReportFatal(ctx context.Context, err error) {
	if l.ReportFatalFn == nil {
		panic("ReportFatalFn not defined")
	}
	l.ReportFatalFn(ctx, err)
}

func (l *LogService) ReportInfo(ctx context.Context, info string) {
	if l.ReportInfoFn == nil {
		panic("ReportInfoFn not defined")
	}
	l.ReportInfoFn(ctx, info)
}

func (l *LogService) ReportPanic(ctx context.Context, err interface{}) {
	if l.ReportPanicFn == nil {
		panic("ReportPanicFn not defined")
	}
	l.ReportPanicFn(ctx, err)
}

func (l *LogService) ReportWarning(ctx context.Context, warning string) {
	if l.ReportWarningFn == nil {
		panic("ReportWarningFn not defined")
	}
	l.ReportWarningFn(ctx, warning)
}

type LogServiceNoOp struct{}

func (l *LogServiceNoOp) Format() string {
	return ""
}

func (l *LogServiceNoOp) Level() int {
	return 0
}

func (l *LogServiceNoOp) Output() io.Writer {
	return io.Discard
}

func (l *LogServiceNoOp) ReportDebug(ctx context.Context, msg string)       {}
func (l *LogServiceNoOp) ReportError(ctx context.Context, err error)        {}
func (l *LogServiceNoOp) ReportFatal(ctx context.Context, err error)        {}
func (l *LogServiceNoOp) ReportInfo(ctx context.Context, info string)       {}
func (l *LogServiceNoOp) ReportPanic(ctx context.Context, err interface{})  {}
func (l *LogServiceNoOp) ReportWarning(ctx context.Context, warning string) {}
