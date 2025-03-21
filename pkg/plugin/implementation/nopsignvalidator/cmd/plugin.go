package main

import (
	"context"

	"github.com/ashishGuliya/onix/pkg/log"
	"github.com/ashishGuliya/onix/pkg/plugin/definition"
)

type provider struct{}

func (vp provider) New(ctx context.Context, config map[string]string) (definition.SignValidator, error) {
	return &defaultValidator{}, nil
}

type defaultValidator struct {
}

func (v *defaultValidator) Verify(ctx context.Context, body []byte, header string, publicKeyBase64 string) (bool, error) {
	log.Debugf(ctx, "NOP Sign Validator called, Skipping sign validation.")
	return true, nil
}

var Provider = provider{}
