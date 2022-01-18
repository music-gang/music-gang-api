package log_test

import (
	"context"
	"io"
	"testing"

	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/service"
	"github.com/music-gang/music-gang-api/log"
	"github.com/slack-go/slack"
)

func TestSlack(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		logService := log.NewSlackLogger("test-webhook-url")

		logService.Lvl = service.LevelAll

		if logService.Output() != io.Discard {
			t.Errorf("Output() = %v, want %v", logService.Output(), io.Discard)
		}

		ctx := context.Background()

		logService.ReportDebug(ctx, "test")
		logService.ReportInfo(ctx, "test")
		logService.ReportWarning(ctx, "test")
		logService.ReportError(ctx, apperr.Errorf(apperr.EINTERNAL, "test"))
		logService.ReportFatal(ctx, apperr.Errorf(apperr.EINTERNAL, "test"))

		logService.Lvl = service.LevelPanic

		func() {

			defer func() {
				if r := recover(); r != nil {
					logService.ReportPanic(ctx, r)
				}
			}()

			panic("test")
		}()
	})

	t.Run("ErrSendWebhook", func(t *testing.T) {

		logService := log.NewSlackLogger("test-webhook-url")
		logServiceOK := log.NewStdOutputLogger()

		logService.Lvl = service.LevelAll
		logService.Fallback = logServiceOK

		logServiceOK.SetOut(io.Discard)

		ctx := context.Background()

		log.PostWebhook = func(url string, msg *slack.WebhookMessage) error {
			return apperr.Errorf(apperr.EINTERNAL, "test")
		}

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
}
