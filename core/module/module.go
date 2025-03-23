package module

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ashishGuliya/onix/core/module/handler"
	"github.com/ashishGuliya/onix/pkg/log"
	"github.com/ashishGuliya/onix/pkg/plugin"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Config struct {
	Name    string `yaml:"name"`
	Path    string `yaml:"path"`
	Handler handler.Config
}

type handlerProvider func(ctx context.Context, mgr *plugin.Manager, cfg *handler.Config) (http.Handler, error)

var handlerProviders = map[handler.HandlerType]handlerProvider{
	handler.HandlerTypeStd:    handler.NewStdHandler,
	handler.HandlerTypeRegSub: handler.NewRegSubscibeHandler,
	handler.HandlerTypeNPSub:  handler.NewNPSubscibeHandler,
	handler.HandlerTypeLookup: handler.NewLookHandler,
}

// AddHandlers registers the handlers for the application.
func Register(ctx context.Context, mCfgs []Config, mux *http.ServeMux, mgr *plugin.Manager) error {
	log.Debugf(ctx, "Registering modules with config: %#v", mCfgs)
	// Iterate over the handlers in the configuration.
	for _, c := range mCfgs {
		rmp, ok := handlerProviders[c.Handler.Type]
		if !ok {
			return fmt.Errorf("invalid module : %s", c.Name)
		}
		h, err := rmp(ctx, mgr, &c.Handler)
		if err != nil {
			return fmt.Errorf("%s : %w", c.Name, err)
		}
		if len(c.Handler.Trace) != 0 {
			h = otelhttp.NewHandler(mux, c.Name)
		}
		h, err = addMiddleware(ctx, mgr, h, &c.Handler)
		if err != nil {
			return fmt.Errorf("failed to add middleware: %w", err)

		}
		log.Debugf(ctx, "Registering handler %s, of type %s @ %s", c.Name, c.Handler.Type, c.Path)
		mux.Handle(c.Path, h)
	}
	return nil
}

// addMiddleware applies middleware to a handler in reverse order.
func addMiddleware(ctx context.Context, mgr *plugin.Manager, handler http.Handler, hCfg *handler.Config) (http.Handler, error) {
	mws := hCfg.Plugins.Middleware
	log.Debugf(ctx, "Applying %d middleware(s) to the handler", len(mws))
	// Apply the middleware in reverse order.
	for i := len(mws) - 1; i >= 0; i-- {
		log.Debugf(ctx, "Loading middleware: %s", mws[i].ID)
		mw, err := mgr.Middleware(ctx, &mws[i])
		if err != nil {
			log.Errorf(ctx, err, "Failed to load middleware %s: %v", mws[i].ID, err)
			return nil, fmt.Errorf("failed to load middleware %s: %w", mws[i].ID, err)
		}
		// Apply the middleware to the handler.
		handler = mw(handler)
		if hCfg.Trace[mws[i].ID] {
			handler = tracingWrapper(mws[i].ID, "middleware", handler)
		}
		log.Debugf(ctx, "Applied middleware: %s", mws[i].ID)
	}

	log.Debugf(ctx, "Middleware chain setup completed")
	return handler, nil
}
