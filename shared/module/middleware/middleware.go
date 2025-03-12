package middleware

import (
	"context"
	"io"
	"net/http"
	"net/http/httputil"

	"onix/shared/log"
	"onix/shared/plugin"
	"onix/shared/plugin/definition"
)

// addMiddleware adds middleware to a handler.
func Chain(ctx context.Context, mgr *plugin.Manager, handler http.Handler, mws []plugin.Config) (http.Handler, error) {
	// Apply the middleware in reverse order.
	for i := len(mws) - 1; i >= 0; i-- {
		// Get the middleware from the plugin manager.
		mw, err := mgr.Middleware(ctx, &mws[i])
		if err != nil {
			return nil, err
		}
		// Apply the middleware to the handler.
		handler = mw(handler)
	}
	// Return the modified handler.
	return handler, nil
}

// ReverseProxy handles incoming HTTP requests and proxies them to the destination.
func ReverseProxy() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		proxy := httputil.NewSingleHostReverseProxy(r.URL)
		log.Debugf(r.Context(), "Proxying request to: %s", r.URL.String())
		// Apply the authentication middleware to the proxy handler.
		proxy.ServeHTTP(w, r)
	}
}

// ProxyHandler handles incoming HTTP requests and proxies them to the destination.
func MsgPublisher(publisher definition.Publisher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		msg, err := io.ReadAll(r.Body)
		if err != nil {
			log.Errorf(r.Context(), err, "failed to read request body")
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		if err := publisher.Publish(r.Context(), msg); err != nil {
			http.Error(w, "Error publishing message", http.StatusInternalServerError)
			return
		}
	}
}
