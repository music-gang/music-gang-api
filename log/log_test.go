package log_test

import (
	"context"
	"io"
	"testing"

	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/service"
	"github.com/music-gang/music-gang-api/log"
	"github.com/music-gang/music-gang-api/mock"
)

var (
	testFormat = service.FormatStandard
	testLevel  = service.LevelAll + 1
)

func TestLog(t *testing.T) {

	t.Run("OK", func(t *testing.T) {
		ctx := context.Background()
		logger := &log.Logger{}

		mockLogger := &mock.LogService{
			LevelFn: func() int {
				return testLevel
			},
			FormatFn: func() string {
				return testFormat
			},
			OutputFn: func() io.Writer {
				return io.Discard
			},
			ReportDebugFn:   func(ctx context.Context, msg string) {},
			ReportErrorFn:   func(ctx context.Context, err error) {},
			ReportFatalFn:   func(ctx context.Context, err error) {},
			ReportInfoFn:    func(ctx context.Context, info string) {},
			ReportPanicFn:   func(ctx context.Context, err interface{}) {},
			ReportWarningFn: func(ctx context.Context, warning string) {},
		}

		logger.AddBackend(mockLogger)

		if logger.Format() != testFormat {
			t.Errorf("logger.Format() = %v, want %v", logger.Format(), testFormat)
		}

		if logger.Level() != testLevel {
			t.Errorf("logger.Level() = %v, want %v", logger.Level(), testLevel)
		}

		if logger.Output() != io.Discard {
			t.Errorf("logger.Output() = %v, want %v", logger.Output(), io.Discard)
		}

		logger.ReportDebug(ctx, "test")
		logger.ReportInfo(ctx, "test")
		logger.ReportWarning(ctx, "test")
		logger.ReportError(ctx, apperr.Errorf(apperr.EINTERNAL, "test"))
		logger.ReportFatal(ctx, apperr.Errorf(apperr.EINTERNAL, "test"))

		func() {

			defer func() {
				if r := recover(); r != nil {
					logger.ReportPanic(ctx, r)
				}
			}()

			panic("test")
		}()
	})
}
