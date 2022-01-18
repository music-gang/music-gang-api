package log_test

import (
	"context"
	"io"
	"testing"

	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/service"
	"github.com/music-gang/music-gang-api/log"
)

func TestStdOutput(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		logService := log.NewStdOutputLogger()

		logService.SetOut(io.Discard)

		if logService.Output() != io.Discard {
			t.Errorf("Output() = %v, want %v", logService.Output(), io.Discard)
		}

		ctx := context.Background()

		logService.ReportDebug(ctx, "test")
		logService.ReportInfo(ctx, "test")
		logService.ReportWarning(ctx, "test")
		logService.ReportError(ctx, apperr.Errorf(apperr.EINTERNAL, "test"))
		logService.ReportFatal(ctx, apperr.Errorf(apperr.EINTERNAL, "test"))

		func() {

			defer func() {
				if r := recover(); r != nil {
					logService.ReportPanic(ctx, r)
				}
			}()

			panic("test")
		}()
	})

	t.Run("WithConfig", func(t *testing.T) {

		t.Run("CustomLevel", func(t *testing.T) {

			logService := log.NewStdOutputLoggerWithConfig(log.StdOutputLoggerOptions{
				Level: service.LevelOff,
			})

			if logService.Level() != service.LevelOff {
				t.Errorf("Expected level %d, got %d", service.LevelOff, logService.Level())
			}

			logService.SetOut(io.Discard)

			ctx := context.Background()

			logService.ReportDebug(ctx, "test")
			logService.ReportInfo(ctx, "test")
			logService.ReportWarning(ctx, "test")
			logService.ReportError(ctx, apperr.Errorf(apperr.EINTERNAL, "test"))
			logService.ReportFatal(ctx, apperr.Errorf(apperr.EINTERNAL, "test"))

			func() {

				defer func() {
					if r := recover(); r != nil {
						logService.ReportPanic(ctx, r)
					}
				}()

				panic("test")
			}()
		})

		t.Run("CustomFormat", func(t *testing.T) {

			logService := log.NewStdOutputLoggerWithConfig(log.StdOutputLoggerOptions{
				Format: service.FormatJSON,
			})

			if logService.Format() != service.FormatJSON {
				t.Errorf("Expected format %s, got %s", service.FormatJSON, logService.Format())
			}

			logService.SetOut(io.Discard)

			ctx := context.Background()

			logService.ReportDebug(ctx, "test")
			logService.ReportInfo(ctx, "test")
			logService.ReportWarning(ctx, "test")
			logService.ReportError(ctx, apperr.Errorf(apperr.EINTERNAL, "test"))
			logService.ReportFatal(ctx, apperr.Errorf(apperr.EINTERNAL, "test"))

			func() {

				defer func() {
					if r := recover(); r != nil {
						logService.ReportPanic(ctx, r)
					}
				}()

				panic("test")
			}()
		})
	})
}
