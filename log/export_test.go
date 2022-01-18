package log

import (
	"io"

	"github.com/slack-go/slack"
)

func init() {
	PostWebhook = func(url string, msg *slack.WebhookMessage) error {
		return nil
	}
}

func (s *StdOutputLogger) SetOut(o io.Writer) {
	s.out = o
}
