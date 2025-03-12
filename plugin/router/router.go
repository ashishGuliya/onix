package router

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

type Config struct {
	Routes []route
}

// route struct to define routing rules.
type route struct {
	// Action is one of the matching criteria.
	Action string `yaml:"action"`

	Type string
	// Target is the URL to proxy to if all criteria match.
	Target string `yaml:"target"`
}

type router struct {
	cfg *Config
}

func (r *router) Target(ctx context.Context, rb []byte) (*route, error) {
	var data map[string]json.RawMessage
	if err := json.Unmarshal(rb, &data); err != nil {
		return nil, fmt.Errorf("invalid request body json")
	}
	// Get the "context" field as a RawMessage.
	contextRaw, ok := data["context"]
	if !ok {
		return nil, fmt.Errorf("context field not found")
	}
	// Unmarshal the "context" RawMessage into a map.
	var contextData map[string]interface{}
	if err := json.Unmarshal(contextRaw, &contextData); err != nil {
		return nil, fmt.Errorf("invalid request.context json")
	}

	// Update the TransactionID in the Context.
	rAction, ok := contextData["action"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid request.context.action json")
	}

	for _, route := range r.cfg.Routes {
		if route.Action == rAction {
			return url.Parse(route.Target)
		}
	}
	return nil, fmt.Errorf("unsupported request.action: %v", rAction)
}

func valid(c *Config) error {
	if c == nil {
		return fmt.Errorf("nil config")
	}
	return nil
}

func New(ctx context.Context, url *url.URL, c *Config) (*router, error) {
	if err := valid(c); err != nil {
		return nil, fmt.Errorf("invalid config: %v", err)
	}

	return &router{cfg: c}, nil
}
