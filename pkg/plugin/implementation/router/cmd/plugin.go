package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ashishGuliya/onix/plugin/router"

	"github.com/ashishGuliya/onix/pkg/log"
	"github.com/ashishGuliya/onix/pkg/plugin/definition"
	"gopkg.in/yaml.v2"
)

type routerProvider struct{}

const pathKey = "routingConfigPath"

// config loads and validates the configuration.
func config(ctx context.Context, path string) (*router.Config, error) {
	// Open the configuration file.
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open config file: %w", err)
	}
	defer file.Close()

	// Decode the YAML configuration.
	var cfg router.Config
	if err := yaml.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("could not decode config: %w", err)
	}
	b, _ := json.MarshalIndent(cfg, "", "  ")
	log.Debugf(ctx, "Loaded %s, \n%s", path, string(b))
	return &cfg, nil
}

func (vp routerProvider) New(ctx context.Context, cfg map[string]string) (definition.Router, error) {
	c, err := config(ctx, cfg[pathKey])
	if err != nil {
		return nil, err
	}
	return router.New(ctx, c)
}

var Provider = routerProvider{}
