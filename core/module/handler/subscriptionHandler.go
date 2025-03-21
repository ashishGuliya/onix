package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ashishGuliya/onix/core/module/client"
	"github.com/ashishGuliya/onix/pkg/log"
	"github.com/ashishGuliya/onix/pkg/plugin"
	"github.com/ashishGuliya/onix/pkg/plugin/definition"
	"github.com/ashishGuliya/onix/pkg/protocol"
	"github.com/ashishGuliya/onix/pkg/response"
)

type registryClient interface {
	Subscribe(ctx context.Context, subscription *protocol.Subscription) error
	Lookup(ctx context.Context, subscription *protocol.Subscription) ([]protocol.Subscription, error)
}

// regSubscibeHandler encapsulates the subscription logic.
type npSubscibeHandler struct {
	km      definition.KeyManager
	rClient registryClient
}

// NewRegSubscibeHandler creates a new instance of SubscriptionService.
func NewNPSubscibeHandler(ctx context.Context, mgr *plugin.Manager, cfg *Config) (http.Handler, error) {
	s := &npSubscibeHandler{
		rClient: client.NewRegisteryClient(&client.Config{RegisteryURL: cfg.RegistryURL}),
	}
	// Initialize plugins
	if err := s.initPlugins(ctx, mgr, &cfg.Plugins); err != nil {
		return nil, fmt.Errorf("failed to initialize plugins: %w", err)
	}

	return s, nil
}

// initPlugins initializes required plugins for the processor.
func (h *npSubscibeHandler) initPlugins(ctx context.Context, mgr *plugin.Manager, cfg *pluginCfg) error {
	var err error
	if cfg.Cache == nil {
		return fmt.Errorf("invalid config: Cache missing")
	}
	cache, err := mgr.Cache(ctx, cfg.Cache)
	if err != nil {
		return fmt.Errorf("failed to load cache: %w", err)
	}
	if cfg.KeyManager == nil {
		return fmt.Errorf("invalid config: KeyManager missing")
	}
	if h.km, err = mgr.KeyManager(ctx, cache, h.rClient, cfg.KeyManager); err != nil {
		return fmt.Errorf("failed to load cache: %w", err)
	}
	return nil
}

// ServeHTTP handles incoming subscription requests.
func (h *npSubscibeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Debug(r.Context(), "NP Subscribe handler called.")
	log.Request(r.Context(), r, nil)
	// Ensure the request method is POST
	if r.Method != http.MethodPost {
		resp, _ := response.Nack(r.Context(), response.MethodNotAllowedType, "invalid request method, only POST allowed", []byte{})
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write(resp)
		return
	}
	// Parse request body
	var reqPayload protocol.Subscription
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Errorf(r.Context(), err, "Reg Subscribe handler: Bad Request, could not read body")
		resp, _ := response.Nack(r.Context(), response.InvalidRequestErrorType, "could not read body", []byte{})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
		return
	}
	if err := json.Unmarshal(body, &reqPayload); err != nil {
		resp, _ := response.Nack(r.Context(), response.InvalidRequestErrorType, "invalid request body", body)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
		return
	}

	// Validate subscriber_id
	if reqPayload.SubscriberID == "" {
		resp, _ := response.Nack(r.Context(), response.InvalidRequestErrorType, "missing subscriber_id", body)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
		return
	}
	// Validate subscriber_id
	if reqPayload.URL == "" {
		resp, _ := response.Nack(r.Context(), response.InvalidRequestErrorType, "missing subscriber url", body)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
		return
	}
	keys, err := h.km.GenerateKeyPairs()
	if err != nil {
		log.Errorf(r.Context(), err, "failed to generate keys")
		resp, _ := response.Nack(r.Context(), response.InternalServerErrorType, "Internal Server Error", body)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(resp)
		return
	}
	log.Debugf(r.Context(), "got keys %#v", keys)
	// Create subscription request
	reqData := &protocol.Subscription{
		KeyID:            keys.UniqueKeyID,
		SigningPublicKey: keys.SigningPublic,
		EncrPublicKey:    keys.EncrPublic,
		Subscriber:       reqPayload.Subscriber,
	}

	if err := h.rClient.Subscribe(r.Context(), reqData); err != nil {
		log.Errorf(r.Context(), err, "Call to registery failed")
		resp, _ := response.Nack(r.Context(), response.InternalServerErrorType, "Internal Server Error", body)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(resp)
		return
	}
	if err := h.km.StorePrivateKeys(r.Context(), reqPayload.SubscriberID, keys); err != nil {
		log.Errorf(r.Context(), err, "StorePrivateKeys failed")
		resp, _ := response.Nack(r.Context(), response.InternalServerErrorType, "Internal Server Error", body)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(resp)
		return
	}
	// Forward the response back to the client
	resp, _ := response.Acknowledge(r.Context(), body)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
