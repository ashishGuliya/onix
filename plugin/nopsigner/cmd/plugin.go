package main

import (
	"context"
	"onix/shared/log"
	"onix/shared/plugin/definition"
)

type provider struct{}

func (vp provider) New(ctx context.Context, config map[string]string) (definition.Signer, func(), error) {
	return &signer{}, nil, nil
}

type signer struct {
}

func (v *signer) Sign(ctx context.Context, body []byte, privateKeyBase64 string) (string, error) {
	log.Debugf(ctx, "NOP Signer called, Returing nop sign.")
	return "NOP Sign", nil
}

var Provider = provider{}
