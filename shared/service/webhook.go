package service

import (
	"bytes"
	"context"
	"log"

	"google.golang.org/api/idtoken"
)

type Webhook struct {
	target string
}

func NewWebhook(target string) *Webhook {
	return &Webhook{
		target: target,
	}
}

func (s *Webhook) Process(payload []byte) error {
	client, err := idtoken.NewClient(context.Background(), s.target)
	if err != nil {
		log.Fatalf("idtoken.NewClient: %w", err)
	}
	if _, err := client.Post(s.target, "application/json", bytes.NewBuffer(payload)); err != nil {
		return err
	}
	return nil
}
