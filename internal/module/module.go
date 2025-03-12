package module

import (
	"context"
	"fmt"
	"net/http"

	"onix/shared/log"
	"onix/shared/module/config"
	"onix/shared/module/transaction"
	"onix/shared/plugin"
	"onix/shared/plugin/definition"
)

type Config struct {
	Modules []config.ModuleCfg
}

type MuduleRegisterProvider func(context.Context, *plugin.Manager, *config.ModuleCfg) (http.Handler, error)

var registery = map[string]MuduleRegisterProvider{
	"transactionProcessor":   transaction.RegisterCaller,
	"transactionReciever": transaction.RegisterReciever,
}

// AddHandlers registers the handlers for the application.
func Register(ctx context.Context, cfg *Config, mux *http.ServeMux, mgr *plugin.Manager) error {
	log.Debugf(ctx, "Registering modules with config: %#v", cfg)
	// Iterate over the handlers in the configuration.
	for _, c := range cfg.Modules {
		rmp, ok := registery[c.Name]
		if !ok {
			return fmt.Errorf("invalid module : %s", c.Name)
		}
		h, err := rmp(ctx, mgr, &c)
		if err != nil {
			return err
		}
		mux.Handle(c.Path, h)
	}
	return nil
}

type processor struct {
}

type processorCtx struct {
	body  *[]byte
	route *definition.Route
	http.Client
}

