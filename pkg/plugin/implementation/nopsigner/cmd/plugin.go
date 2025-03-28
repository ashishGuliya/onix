package main

import (
	"context"

	"github.com/ashishGuliya/onix/pkg/log"
	"github.com/ashishGuliya/onix/pkg/plugin/definition"
)

type provider struct{}

func (vp provider) New(ctx context.Context, config map[string]string) (definition.Signer, func(), error) {
	return &signer{}, nil, nil
}

type signer struct {
}

func (v *signer) Sign(ctx context.Context, _ []byte, _ string, _ int64, _ int64) (string, error) {
	log.Debugf(ctx, "NOP Signer called, Returing nop sign.")
	return "NOP Sign", nil
}

var Provider = provider{}
