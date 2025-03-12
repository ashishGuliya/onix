package processor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/ashishGuliya/onix/internal/module/config"
	"github.com/ashishGuliya/onix/pkg/log"
	"github.com/ashishGuliya/onix/pkg/plugin"
	"github.com/ashishGuliya/onix/pkg/plugin/definition"
)

// processor orchestrates the execution of defined processing steps.
type processor struct {
	signer          definition.Signer
	steps           []definition.Step
	signValidator   definition.SignValidator
	schemaValidator definition.SchemaValidator
	router          definition.Router
	publisher       definition.Publisher
}

// NewProcessor initializes a new processor with plugins and steps.
func NewProcessor(ctx context.Context, mgr *plugin.Manager, cfg *config.ModuleCfg) (*processor, error) {
	p := &processor{
		steps: []definition.Step{},
	}

	// Initialize plugins
	if err := p.initPlugins(ctx, mgr, &cfg.Plugins); err != nil {
		return nil, fmt.Errorf("failed to initialize plugins: %w", err)
	}

	// Initialize steps
	if err := p.initSteps(ctx, mgr, cfg); err != nil {
		return nil, fmt.Errorf("failed to initialize steps: %w", err)
	}

	return p, nil
}

// Process executes defined processing steps on an incoming request.
func (p *processor) Process(w http.ResponseWriter, r *http.Request) {
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
	route(ctx, r, w, p)
}

func route(ctx *definition.StepContext, r *http.Request, w http.ResponseWriter, p *processor, pb *definition.Publisher) {
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

// initPlugins initializes required plugins for the processor.
func (p *processor) initPlugins(ctx context.Context, mgr *plugin.Manager, cfg *config.PluginCfg) error {
	var err error

	if cfg.SignValidator != nil {
		if p.signValidator, err = mgr.SignValidator(ctx, cfg.SignValidator); err != nil {
			return fmt.Errorf("failed to load sign validator: %w", err)
		}
	}

	if cfg.SchemaValidator != nil {
		if p.schemaValidator, err = mgr.Validator(ctx, cfg.SchemaValidator); err != nil {
			return fmt.Errorf("failed to load schema validator: %w", err)
		}
	}

	if cfg.Router != nil {
		if p.router, err = mgr.Router(ctx, cfg.Router); err != nil {
			return fmt.Errorf("failed to load router: %w", err)
		}
	}

	if cfg.Publisher != nil {
		if p.publisher, err = mgr.Publisher(ctx, cfg.Publisher); err != nil {
			return fmt.Errorf("failed to load publisher: %w", err)
		}
	}

	if cfg.Signer != nil {
		if p.signer, err = mgr.Signer(ctx, cfg.Signer); err != nil {
			return fmt.Errorf("failed to load signer: %w", err)
		}
	}

	return nil
}

// initSteps initializes and validates processing steps for the processor.
func (p *processor) initSteps(ctx context.Context, mgr *plugin.Manager, cfg *config.ModuleCfg) error {
	steps := make(map[string]definition.Step)

	// Validate plugin dependencies before proceeding
	if err := validateStepDependencies(cfg, p); err != nil {
		return err
	}

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
		switch step {
		case "Sign":
			p.steps = append(p.steps, &signStep{signer: p.signer})
		case "ValidateSign":
			p.steps = append(p.steps, &validateSignStep{validator: p.signValidator})
		case "ValidateSchema":
			p.steps = append(p.steps, &validateSchemaStep{validator: p.schemaValidator})
		case "Broadcast":
			p.steps = append(p.steps, &broadcastStep{})
		case "GetRoute":
			p.steps = append(p.steps, &getRouteStep{router: p.router})
		default:
			if customStep, exists := steps[step]; exists {
				p.steps = append(p.steps, customStep)
			} else {
				return fmt.Errorf("unrecognized step: %s", step)
			}
		}
	}

	log.Infof(ctx, "Processor steps initialized: %v", cfg.Steps)
	return nil
}

// validateStepDependencies ensures required plugins are loaded for configured steps.
func validateStepDependencies(cfg *config.ModuleCfg, p *processor) error {
	if contains(cfg.Steps, "Sign") && p.signer == nil {
		return fmt.Errorf("invalid config: Signer plugin not configured")
	}
	if contains(cfg.Steps, "ValidateSign") && p.signValidator == nil {
		return fmt.Errorf("invalid config: SignValidator plugin not configured")
	}
	if contains(cfg.Steps, "ValidateSchema") && p.schemaValidator == nil {
		return fmt.Errorf("invalid config: SchemaValidator plugin not configured")
	}
	if contains(cfg.Steps, "GetRoute") && p.router == nil {
		return fmt.Errorf("invalid config: Router plugin not configured")
	}
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
