package main

import (
	"context"
	"errors"

	"github.com/ashishGuliya/onix/pkg/plugin/definition"
	"github.com/ashishGuliya/onix/pkg/plugin/implementation/signer"
)

// provider implements the definition.provider interface.
type signerProvider struct{}

// New creates a new Signer instance using the provided configuration.
func (p signerProvider) New(ctx context.Context, config map[string]string) (definition.Signer, func() error, error) {
	if ctx == nil {
		return nil, nil, errors.New("context cannot be nil")
	}
	return signer.New()
}

// Provider is the exported symbol that the plugin manager will look for.
var Provider = signerProvider{}
