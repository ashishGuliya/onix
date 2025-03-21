package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ashishGuliya/onix/pkg/log"
	"github.com/ashishGuliya/onix/pkg/plugin"
	"github.com/ashishGuliya/onix/pkg/plugin/definition"
	"github.com/ashishGuliya/onix/pkg/protocol"
	"github.com/ashishGuliya/onix/pkg/response"
)

// regSubscibeHandler encapsulates the subscription logic.
type regSubscibeHandler struct {
	cache definition.Cache
}

// NewRegSubscibeHandler creates a new instance of SubscriptionService.
func NewRegSubscibeHandler(ctx context.Context, mgr *plugin.Manager, cfg *Config) (http.Handler, error) {
	s := &regSubscibeHandler{}
	// Initialize plugins
	if err := s.initPlugins(ctx, mgr, &cfg.Plugins); err != nil {
		return nil, fmt.Errorf("failed to initialize plugins: %w", err)
	}
	return s, nil
}

// initPlugins initializes required plugins for the processor.
func (p *regSubscibeHandler) initPlugins(ctx context.Context, mgr *plugin.Manager, cfg *pluginCfg) error {
	var err error
	if cfg.Cache == nil {
		return fmt.Errorf("invalid config: Cache missing")
	}
	if p.cache, err = mgr.Cache(ctx, cfg.Cache); err != nil {
		return fmt.Errorf("failed to load cache: %w", err)
	}
	return nil
}

// SubscribeHandler processes subscription requests.
func (s *regSubscibeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Debug(r.Context(), "Reg Subscribe handler called.")

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Errorf(r.Context(), err, "Error reading request body")
		nackResponse, _ := response.Nack(r.Context(), response.InvalidRequestErrorType, "Error reading request body", []byte{})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(nackResponse)
		return
	}

	var req protocol.Subscription
	if err2 := json.Unmarshal(bodyBytes, &req); err2 != nil {
		log.Errorf(r.Context(), err2, "Reg Subscribe handler: Bad Request")
		nackResponse, _ := response.Nack(r.Context(), response.InvalidRequestErrorType, err2.Error(), bodyBytes)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(nackResponse)
		return
	}

	// Validate the request
	if err := validateSubscriptionReq(&req); err != nil {
		log.Errorf(r.Context(), err, "Reg Subscribe handler: Bad Request")
		nackResponse, _ := response.Nack(r.Context(), response.InvalidRequestErrorType, err.Error(), bodyBytes)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(nackResponse)
		return
	}

	// Process subscription
	if err := s.subscribe(&req); err != nil {
		log.Errorf(r.Context(), err, "failed to process subscription")
		nackResponse, _ := response.Nack(r.Context(), response.InternalServerErrorType, "failed to process subscription", bodyBytes)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(nackResponse)
		return
	}

	// Respond with success
	ackResponse, _ := response.Acknowledge(r.Context(), bodyBytes)
	w.WriteHeader(http.StatusOK)
	w.Write(ackResponse)
}

// validate checks if all required fields are present and valid.
func validateSubscriptionReq(req *protocol.Subscription) error {
	if req == nil {
		return errors.New("missing request")
	}
	if req.SigningPublicKey == "" {
		return errors.New("missing signing public key")
	}
	if req.EncrPublicKey == "" {
		return errors.New("missing encryption public key")
	}
	if req.URL == "" {
		return errors.New("missing URL")
	}
	return nil
}

// subscribe processes the subscription logic and stores it in the cache.
func (s *regSubscibeHandler) subscribe(req *protocol.Subscription) error {
	subscription := &protocol.Subscription{
		Subscriber:       req.Subscriber,
		SigningPublicKey: req.SigningPublicKey,
		EncrPublicKey:    req.EncrPublicKey,
		KeyID:            req.KeyID,
		Status:           "UNDER_SUBSCRIPTION",
		ValidFrom:        time.Now(),
		ValidUntil:       time.Now().Add(48 * time.Hour),
		Created:          time.Now(),
		Updated:          time.Now(),
	}

	// Store in cache
	cacheKey := fmt.Sprintf("subscriber:%s", req.SubscriberID)
	subscriptionData, err := json.Marshal(subscription)
	if err != nil {
		return fmt.Errorf("failed to marshal subscription data: %w", err)
	}
	return s.cache.Set(context.Background(), cacheKey, string(subscriptionData), 240*time.Hour) // Default 24hr TTL
}

// lookUpHandler encapsulates the lookup logic.
type lookUpHandler struct {
	cache definition.Cache
}

// NewLookHandler creates a new instance of RegistryHandler.
func NewLookHandler(ctx context.Context, mgr *plugin.Manager, cfg *Config) (http.Handler, error) {
	h := &lookUpHandler{}
	if err := h.initPlugins(ctx, mgr, &cfg.Plugins); err != nil {
		return nil, fmt.Errorf("failed to initialize plugins: %w", err)
	}
	return h, nil
}

// initPlugins initializes required plugins for the processor.
func (h *lookUpHandler) initPlugins(ctx context.Context, mgr *plugin.Manager, cfg *pluginCfg) error {
	var err error
	if cfg.Cache == nil {
		return fmt.Errorf("invalid config: Cache missing")
	}
	if h.cache, err = mgr.Cache(ctx, cfg.Cache); err != nil {
		return fmt.Errorf("failed to load cache: %w", err)
	}
	return nil
}

// LookupHandler handles the lookup requests.
func (h *lookUpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		nackResponse, _ := response.Nack(r.Context(), response.MethodNotAllowedType, "Method Not Allowed", []byte{})
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write(nackResponse)
		return
	}
	log.Debug(r.Context(), "Reg Lookup handler called.")

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Errorf(r.Context(), err, "Error reading request body")
		nackResponse, _ := response.Nack(r.Context(), response.InvalidRequestErrorType, "Error reading request body", []byte{})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(nackResponse)
		return
	}

	var req protocol.Subscription
	if err2 := json.Unmarshal(bodyBytes, &req); err2 != nil {
		nackResponse, _ := response.Nack(r.Context(), response.InvalidRequestErrorType, err2.Error(), bodyBytes)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(nackResponse)
		return
	}

	ctx := context.Background()

	// Use the SubscriberID as the cache key
	cacheKey := fmt.Sprintf("subscriber:%s", req.SubscriberID)

	cachedValue, err := h.cache.Get(ctx, cacheKey)
	if err != nil {
		nackResponse, _ := response.Nack(r.Context(), response.NotFoundType, "Subscriber ID not found", bodyBytes)
		w.WriteHeader(http.StatusNotFound)
		w.Write(nackResponse)
		return
	}

	var subData protocol.Subscription
	err = json.Unmarshal([]byte(cachedValue), &subData)
	if err != nil {
		nackResponse, _ := response.Nack(r.Context(), response.InternalServerErrorType, "Internal Server Error", bodyBytes)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(nackResponse)
		log.Errorf(r.Context(), err, "Error unmarshaling cached data")
		return
	}

	// Send the result as the response
	w.Header().Set("Content-Type", "application/json")
	ackResponse, err3 := json.Marshal(response.BecknResponse{Context: map[string]interface{}{}, Message: response.Message{Ack: struct {
		Status string `json:"status,omitempty"`
	}{Status: "ACK"}}})
	if err3 != nil {
		log.Errorf(r.Context(), err3, "error creating ack response")
	}
	w.Write(ackResponse)
	err = json.NewEncoder(w).Encode([]protocol.Subscription{subData})
	if err != nil {
		log.Errorf(r.Context(), err, "Error encoding JSON")
	}
}
