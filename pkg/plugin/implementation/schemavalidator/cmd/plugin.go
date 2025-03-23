package main

import (
	"context"
	"errors"

	"github.com/ashishGuliya/onix/pkg/plugin/definition"
	"github.com/ashishGuliya/onix/pkg/plugin/implementation/schemavalidator"
)

// schemaValidatorProvider provides instances of schemaValidator.
type schemaValidatorProvider struct{}

// New initializes a new Verifier instance.
func (vp schemaValidatorProvider) New(ctx context.Context, config map[string]string) (definition.SchemaValidator, func() error, error) {
	if ctx == nil {
		return nil, nil, errors.New("context cannot be nil")
	}

	// Extract schema_dir from the config map
	schemaDir, ok := config["schemaDir"]
	if !ok || schemaDir == "" {
		return nil, nil, errors.New("config must contain 'schemaDir'")
	}

	// Create a new schemaValidator instance with the provided configuration
	return schemavalidator.New(ctx, &schemavalidator.Config{
		SchemaDir: schemaDir,
	})
}

var Provider = schemaValidatorProvider{}
