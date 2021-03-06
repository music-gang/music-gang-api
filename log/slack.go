package log

import (
	"context"
	"encoding/json"
	"io"
	"strconv"
	"time"

	"github.com/music-gang/music-gang-api/app"
	"github.com/music-gang/music-gang-api/app/service"
	"github.com/slack-go/slack"
)

const (
	slackDefaultColor = "Yellow"
	slackDefaultTitle = "Slack Logger"
)

var colorMapper = map[int]string{
	service.LevelError: "Yellow",
	service.LevelFatal: "Red",
	service.LevelPanic: "Black",
}

// PostWebhookWithContext posts a message to the Slack webhook.
// It is defined as a separate function to allow mocking in tests.
var PostWebhookWithContext func(ctx context.Context, url string, msg *slack.WebhookMessage) error = slack.PostWebhookContext

var _ service.LogService = (*SlackLogger)(nil)

// SlackLogger is a logger that sends messages to a Slack channel webhook.
type SlackLogger struct {
	Webhook string
	Fmt     string
	Lvl     int

	// Fallback is the logger to use when the webhook fails.
	// I hope you don't use SlackLogger itself as a fallback. If you know what i mean ( ͡° ͜ʖ ͡°)
	Fallback service.LogService
}

// NewSlackLogger creates a new SlackLogger.
// By default, the logger will send messages with log level "error" and above with service.FormatStandard.
func NewSlackLogger(webhook string) *SlackLogger {
	return &SlackLogger{
		Webhook: webhook,
		Fmt:     service.FormatStandard,
		Lvl:     service.LevelError,
	}
}

// Format return the format of the logger.
func (s *SlackLogger) Format() string {
	return s.Fmt
}

// Level returns the current log level.
func (s *SlackLogger) Level() int {
	return s.Lvl
}

// Output returns a writer that writes to the logger.
func (s *SlackLogger) Output() io.Writer {
	// In this case, we don't need to return an io.Writer.
	return io.Discard
}

// ReportDebug logs a message at level Debug.
func (s *SlackLogger) ReportDebug(ctx context.Context, msg string) {
	if s.Level() <= service.LevelDebug {
		if err := s.sendToWebhook(ctx, service.FormatOutputForReportFunc(service.LevelDebug, msg, app.TagsFromContext(ctx), s.Format()), service.LevelDebug); err != nil && s.Fallback != nil {
			// error dial with slack webhook but defined a fallback log service
			s.Fallback.ReportDebug(ctx, msg)
		}
	}
}

// ReportError logs an error.
func (s *SlackLogger) ReportError(ctx context.Context, err error) {
	if s.Level() <= service.LevelError {
		if err := s.sendToWebhook(ctx, service.FormatOutputForReportFunc(service.LevelError, err, app.TagsFromContext(ctx), s.Format()), service.LevelError); err != nil && s.Fallback != nil {
			// error dial with slack webhook but defined a fallback log service
			s.Fallback.ReportError(ctx, err)
		}
	}
}

// ReportFatal logs a fatal error.
func (s *SlackLogger) ReportFatal(ctx context.Context, err error) {
	if s.Level() <= service.LevelFatal {
		if err := s.sendToWebhook(ctx, service.FormatOutputForReportFunc(service.LevelFatal, err, app.TagsFromContext(ctx), s.Format()), service.LevelFatal); err != nil && s.Fallback != nil {
			// error dial with slack webhook but defined a fallback log service
			s.Fallback.ReportFatal(ctx, err)
		}
	}
}

// ReportInfo logs an info.
func (s *SlackLogger) ReportInfo(ctx context.Context, info string) {
	if s.Level() <= service.LevelInfo {
		if err := s.sendToWebhook(ctx, service.FormatOutputForReportFunc(service.LevelInfo, info, app.TagsFromContext(ctx), s.Format()), service.LevelInfo); err != nil && s.Fallback != nil {
			// error dial with slack webhook but defined a fallback log service
			s.Fallback.ReportInfo(ctx, info)
		}
	}
}

// ReportPanic logs a panic.
func (s *SlackLogger) ReportPanic(ctx context.Context, err interface{}) {
	if s.Level() <= service.LevelPanic {
		if err := s.sendToWebhook(ctx, service.FormatOutputForReportFunc(service.LevelPanic, err, app.TagsFromContext(ctx), s.Format()), service.LevelPanic); err != nil && s.Fallback != nil {
			// error dial with slack webhook but defined a fallback log service
			s.Fallback.ReportPanic(ctx, err)
		}
	}
}

// ReportWarning logs a warning.
func (s *SlackLogger) ReportWarning(ctx context.Context, warning string) {
	if s.Level() <= service.LevelWarn {
		if err := s.sendToWebhook(ctx, service.FormatOutputForReportFunc(service.LevelWarn, warning, app.TagsFromContext(ctx), s.Format()), service.LevelWarn); err != nil && s.Fallback != nil {
			// error dial with slack webhook but defined a fallback log service
			s.Fallback.ReportWarning(ctx, warning)
		}
	}
}

// sendToWebhook sends a message to the Slack webhook.
func (s *SlackLogger) sendToWebhook(ctx context.Context, message string, logLevel int) error {

	color := slackDefaultColor
	if level, ok := colorMapper[logLevel]; ok {
		color = level
	}

	attachment := slack.Attachment{
		Color:      color,
		AuthorName: slackDefaultTitle,
		Text:       message,
		Ts:         json.Number(strconv.FormatInt(time.Now().Unix(), 10)),
	}
	msg := &slack.WebhookMessage{
		Attachments: []slack.Attachment{attachment},
	}

	err := PostWebhookWithContext(ctx, s.Webhook, msg)
	if err != nil {
		return err
	}

	return nil
}
