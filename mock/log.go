package mock

import "github.com/inconshreveable/log15"

var _ log15.Logger = (*Logger)(nil)

type Logger struct {
	CritFn       func(msg string, ctx ...interface{})
	DebugFn      func(msg string, ctx ...interface{})
	ErrorFn      func(msg string, ctx ...interface{})
	GetHandlerFn func() log15.Handler
	InfoFn       func(msg string, ctx ...interface{})
	NewFn        func(ctx ...interface{}) log15.Logger
	SetHandlerFn func(h log15.Handler)
	WarnFn       func(msg string, ctx ...interface{})
}

func (l *Logger) Crit(msg string, ctx ...interface{}) {
	if l.CritFn == nil {
		panic("CritFn not defined")
	}
	l.CritFn(msg, ctx...)
}

func (l *Logger) Debug(msg string, ctx ...interface{}) {
	if l.DebugFn == nil {
		panic("DebugFn not defined")
	}
	l.DebugFn(msg, ctx...)
}

func (l *Logger) Error(msg string, ctx ...interface{}) {
	if l.ErrorFn == nil {
		panic("ErrorFn not defined")
	}
	l.ErrorFn(msg, ctx...)
}

func (l *Logger) GetHandler() log15.Handler {
	if l.GetHandlerFn == nil {
		panic("GetHandlerFn not defined")
	}
	return l.GetHandlerFn()
}

func (l *Logger) Info(msg string, ctx ...interface{}) {
	if l.InfoFn == nil {
		panic("InfoFn not defined")
	}
	l.InfoFn(msg, ctx...)
}

func (l *Logger) New(ctx ...interface{}) log15.Logger {
	if l.NewFn == nil {
		panic("NewFn not defined")
	}
	return l.NewFn(ctx...)
}

func (l *Logger) SetHandler(h log15.Handler) {
	if l.SetHandlerFn == nil {
		panic("SetHandlerFn not defined")
	}
	l.SetHandlerFn(h)
}

func (l *Logger) Warn(msg string, ctx ...interface{}) {
	if l.WarnFn == nil {
		panic("WarnFn not defined")
	}
	l.WarnFn(msg, ctx...)
}

var _ log15.Logger = (*LoggerNoOp)(nil)

type LoggerNoOp struct {
	CritFn  func(msg string, ctx ...interface{})
	DebugFn func(msg string, ctx ...interface{})
	ErrorFn func(msg string, ctx ...interface{})
	InfoFn  func(msg string, ctx ...interface{})
	WarnFn  func(msg string, ctx ...interface{})
}

func (l *LoggerNoOp) Crit(msg string, ctx ...interface{}) {
	if l.CritFn != nil {
		l.CritFn(msg, ctx...)
	}
}

func (l *LoggerNoOp) Debug(msg string, ctx ...interface{}) {
	if l.DebugFn != nil {
		l.DebugFn(msg, ctx...)
	}
}

func (l *LoggerNoOp) Error(msg string, ctx ...interface{}) {
	if l.ErrorFn != nil {
		l.ErrorFn(msg, ctx...)
	}
}

func (l *LoggerNoOp) GetHandler() log15.Handler {
	return nil
}

func (l *LoggerNoOp) Info(msg string, ctx ...interface{}) {
	if l.InfoFn != nil {
		l.InfoFn(msg, ctx...)
	}
}

func (l *LoggerNoOp) New(ctx ...interface{}) log15.Logger {
	return &LoggerNoOp{}
}

func (l *LoggerNoOp) SetHandler(h log15.Handler) {}

func (l *LoggerNoOp) Warn(msg string, ctx ...interface{}) {
	if l.WarnFn != nil {
		l.WarnFn(msg, ctx...)
	}
}
