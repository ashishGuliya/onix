package main

import (
	"context"
	"net/url"

	"github.com/ashishGuliya/onix/pkg/log"
	"github.com/ashishGuliya/onix/pkg/plugin/definition"
)

type provider struct{}

func (vp provider) New(ctx context.Context, config map[string]string) (definition.SchemaValidator, error) {
	return &defaultValidator{}, nil
}

type defaultValidator struct {
}

func (v *defaultValidator) Validate(ctx context.Context, url *url.URL, b []byte) error {
	log.Debugf(ctx, "NOP Schema Validator called, Skipping schema validation.")
	return nil
}

var Provider = provider{}
