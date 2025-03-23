package main

import (
	"context"
	"net/http"

	"github.com/ashishGuliya/onix/pkg/plugin/implementation/gcpAuthMdw"
)

// provider implements the PublisherProvider interface.
type provider struct{}

// New creates a new Publisher instance.
func (p provider) New(ctx context.Context, c map[string]string) (func(http.Handler) http.Handler, error) {
	return gcpAuthMdw.New(ctx, c), nil
}

// Provider is the exported symbol that the plugin manager will look for.
var Provider = provider{}
