package main

import (
	"context"

	"github.com/ashishGuliya/onix/pkg/plugin/definition"
	"github.com/ashishGuliya/onix/plugin/signvalidator"
)

// validatorProvider provides instances of Verifier.
type validatorProvider struct{}

// New initializes a new Verifier instance.
func (vp validatorProvider) New(ctx context.Context, config map[string]string) (definition.SignValidator, func() error, error) {
	return signvalidator.New()
}

// Provider is the exported symbol that the plugin manager will look for.
var Provider = validatorProvider{}
