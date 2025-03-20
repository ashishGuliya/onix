package handler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/ashishGuliya/onix/core/module/client"
	"github.com/ashishGuliya/onix/pkg/log"
	"github.com/ashishGuliya/onix/pkg/plugin"
	"github.com/ashishGuliya/onix/pkg/plugin/definition"
)

// stdHandler orchestrates the execution of defined processing steps.
type stdHandler struct {
	signer          definition.Signer
	steps           []definition.Step
	signValidator   definition.SignValidator
	cache           definition.Cache
	km              definition.KeyManager
	schemaValidator definition.SchemaValidator
	router          definition.Router
	publisher       definition.Publisher
}

// NewStdHandler initializes a new processor with plugins and steps.
func NewStdHandler(ctx context.Context, mgr *plugin.Manager, cfg *Config) (http.Handler, error) {
	p := &stdHandler{
		steps: []definition.Step{},
	}
	// Initialize plugins
	if err := p.initPlugins(ctx, mgr, &cfg.Plugins, cfg.RegistryURL); err != nil {
		return nil, fmt.Errorf("failed to initialize plugins: %w", err)
	}
	// Initialize steps
	if err := p.initSteps(ctx, mgr, cfg); err != nil {
		return nil, fmt.Errorf("failed to initialize steps: %w", err)
	}
	return p, nil
}

// Process executes defined processing steps on an incoming request.
func (p *stdHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Efficiently read the request body into a buffer
	var bodyBuffer bytes.Buffer
	if _, err := io.Copy(&bodyBuffer, r.Body); err != nil {
		log.Errorf(r.Context(), err, "Failed to read request body")
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	r.Body.Close()
	ctx := &definition.StepContext{
		Context: r.Context(),
		Request: r,
		Body:    bodyBuffer.Bytes(),
	}
	log.Request(r.Context(), r, ctx.Body)

	// Execute processing steps
	for _, step := range p.steps {
		if err := step.Run(ctx); err != nil {
			log.Errorf(r.Context(), err, "Step execution failed: %T", step)
			http.Error(w, "Internal error during processing", http.StatusInternalServerError)
			return
		}
	}

	// Restore request body before forwarding or publishing
	r.Body = io.NopCloser(bytes.NewReader(ctx.Body))

	if ctx.Route == nil {
		return
	}

	// Handle routing based on the defined route type
	route(ctx, r, w, p.publisher)
}

func route(ctx *definition.StepContext, r *http.Request, w http.ResponseWriter, pb definition.Publisher) {
	log.Debugf(ctx, "Routing to ctx.Route to %#v", ctx.Route)
	switch ctx.Route.Type {
	case "url":
		log.Infof(ctx.Context, "Forwarding request to URL: %s", ctx.Route.URL)
		proxy(r, w, ctx.Route.URL)
		return
	case "publisher":
		if pb == nil {
			err := fmt.Errorf("publisher plugin not configured")
			log.Errorf(ctx.Context, err, "Invalid configuration")
			http.Error(w, "Invalid configuration: Publisher plugin not configured", http.StatusInternalServerError)
			return
		}
		log.Infof(ctx.Context, "Publishing message to: %s", ctx.Route.Publisher)
		if err := pb.Publish(ctx, ctx.Route.Publisher, ctx.Body); err != nil {
			log.Errorf(ctx.Context, err, "Failed to publish message")
			http.Error(w, "Error publishing message", http.StatusInternalServerError)
			return
		}
	default:
		log.Errorf(ctx.Context, fmt.Errorf("Failed to publish message"), "")
		http.Error(w, "Error publishing message", http.StatusInternalServerError)
		return
	}
}

// proxy forwards the request to a target URL using a reverse proxy.
func proxy(r *http.Request, w http.ResponseWriter, target *url.URL) {
	r.URL.Scheme = target.Scheme
	r.URL.Host = target.Host
	r.URL.Path = target.Path

	r.Header.Set("X-Forwarded-Host", r.Host)
	proxy := httputil.NewSingleHostReverseProxy(target)
	log.Infof(r.Context(), "Proxying request to: %s", target)

	proxy.ServeHTTP(w, r)
}

// loadPlugin is a generic function to load and validate plugins.
func loadPlugin[T any](ctx context.Context, name string, cfg *plugin.Config, mgrFunc func(context.Context, *plugin.Config) (T, error)) (T, error) {
	var zero T
	if cfg == nil {
		log.Debugf(ctx, "Skipping %s plugin: not configured", name)
		return zero, nil
	}

	plugin, err := mgrFunc(ctx, cfg)
	if err != nil {
		return zero, fmt.Errorf("failed to load %s plugin (%s): %w", name, cfg.ID, err)
	}

	log.Debugf(ctx, "Loaded %s plugin: %s", name, cfg.ID)
	return plugin, nil
}

func loadKeyManager(ctx context.Context, mgr *plugin.Manager, cache definition.Cache, cfg *plugin.Config, regURL string) (definition.KeyManager, error) {
	if cfg == nil {
		log.Debug(ctx, "Skipping KeyManager plugin: not configured")
		return nil, nil
	}
	if cache == nil {
		return nil, fmt.Errorf("failed to load KeyManager plugin (%s): Cache plugin not configured", cfg.ID)
	}
	rClient := client.NewRegisteryClient(&client.Config{RegisteryURL: regURL})
	km, err := mgr.KeyManager(ctx, cache, rClient, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load cache plugin (%s): %w", cfg.ID, err)
	}

	log.Debugf(ctx, "Loaded Keymanager plugin: %s", cfg.ID)
	return km, nil
}

// initPlugins initializes required plugins for the processor.
func (p *stdHandler) initPlugins(ctx context.Context, mgr *plugin.Manager, cfg *pluginCfg, regURL string) error {
	var err error
	if p.cache, err = loadPlugin(ctx, "Cache", cfg.Cache, mgr.Cache); err != nil {
		return err
	}
	if p.km, err = loadKeyManager(ctx, mgr, p.cache, cfg.KeyManager, regURL); err != nil {
		return err
	}
	if p.signValidator, err = loadPlugin(ctx, "SignValidator", cfg.SignValidator, mgr.SignValidator); err != nil {
		return err
	}
	if p.schemaValidator, err = loadPlugin(ctx, "SchemaValidator", cfg.SchemaValidator, mgr.SchemaValidator); err != nil {
		return err
	}
	if p.router, err = loadPlugin(ctx, "Router", cfg.Router, mgr.Router); err != nil {
		return err
	}
	if p.publisher, err = loadPlugin(ctx, "Publisher", cfg.Publisher, mgr.Publisher); err != nil {
		return err
	}
	if p.signer, err = loadPlugin(ctx, "Signer", cfg.Signer, mgr.Signer); err != nil {
		return err
	}

	log.Debugf(ctx, "All required plugins successfully loaded for stdHandler")
	return nil
}

// initSteps initializes and validates processing steps for the processor.
func (p *stdHandler) initSteps(ctx context.Context, mgr *plugin.Manager, cfg *Config) error {
	steps := make(map[string]definition.Step)

	// Load plugin-based steps
	for _, c := range cfg.Plugins.Steps {
		step, err := mgr.Step(ctx, &c)
		if err != nil {
			return fmt.Errorf("failed to initialize plugin step %s: %w", c.ID, err)
		}
		steps[c.ID] = step
	}

	// Register processing steps
	for _, step := range cfg.Steps {
		var s definition.Step
		var err error

		switch step {
		case "sign":
			s, err = newSignStep(p.signer, p.km)
		case "validateSign":
			s, err = newValidateSignStep(p.signValidator, p.km)
		case "validateSchema":
			s, err = newValidateSchemaStep(p.schemaValidator)
		case "addRoute":
			s, err = newRouteStep(p.router)
		case "broadcast":
			s = &broadcastStep{}
		default:
			if customStep, exists := steps[step]; exists {
				s = customStep
			} else {
				return fmt.Errorf("unrecognized step: %s", step)
			}
		}

		if err != nil {
			return err
		}

		p.steps = append(p.steps, s)
	}

	log.Infof(ctx, "Processor steps initialized: %v", cfg.Steps)
	return nil
}

// contains checks if a slice contains a given string.
func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}
