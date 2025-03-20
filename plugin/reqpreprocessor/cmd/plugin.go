package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/ashishGuliya/onix/plugin/reqpreprocessor"
)

// provider implements the PublisherProvider interface.
type provider struct{}

// New creates a new Publisher instance.
func (p provider) New(ctx context.Context, c map[string]string) (func(http.Handler) http.Handler, error) {
	config := &reqpreprocessor.Config{}
	if uuidKeysStr, ok := c["uuidKeys"]; ok {
		config.UUIDKeys = strings.Split(uuidKeysStr, ",")
	}

	if role, ok := c["role"]; ok {
		config.Role = role
	}

	return reqpreprocessor.NewUUIDSetter(config)
}

// Provider is the exported symbol that the plugin manager will look for.
var Provider = provider{}
