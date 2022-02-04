package log

import (
	"context"
	"io"

	"github.com/slack-go/slack"
)

func init() {
	PostWebhookWithContext = func(ctx context.Context, url string, msg *slack.WebhookMessage) error {
		return nil
	}
}

func (s *StdOutputLogger) SetOut(o io.Writer) {
	s.out = o
}
