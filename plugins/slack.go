package plugins

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
)

type SimpleSlackNotify struct {
	client  *slack.Client
	channel string
}

func NewSimpleSlackNotify(token string, channel string) *SimpleSlackNotify {
	c := slack.New(token)

	return &SimpleSlackNotify{
		client:  c,
		channel: channel,
	}
}

func (s *SimpleSlackNotify) Run(ctx context.Context, taskID string) error {
	msg := fmt.Sprintf("reindex task [ID=%v] is done", taskID)
	_, _, err := s.client.PostMessage(s.channel, slack.MsgOptionText(msg, true))
	if err != nil {
		return err
	}
	return nil
}
