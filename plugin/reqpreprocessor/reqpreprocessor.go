package reqpreprocessor

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/ashishGuliya/onix/pkg/log"
	"github.com/google/uuid"
)

// Config holds the configuration for the middleware.
type Config struct {
	UUIDKeys []string
	Role     string
}

// contextKey is a private constant for the context key.
const contextKey = "context"

// NewUUIDSetter creates a new request preprocessor middleware that sets UUIDs for the given keys within the context.
func NewUUIDSetter(cfg *Config) (func(http.Handler) http.Handler, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Read the request body.
			var data map[string]any
			if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
				log.Errorf(r.Context(), err, "json.NewDecoder(r.Body): %v", err)
				http.Error(w, "Failed to decode request body", http.StatusBadRequest)
				return
			}

			// Get the context field.
			contextRaw := data[contextKey]
			if contextRaw == nil {
				log.Errorf(r.Context(), fmt.Errorf("%s field not found", contextKey), "")
				http.Error(w, fmt.Sprintf("%s field not found.", contextKey), http.StatusBadRequest)
				return
			}

			// Unmarshal the context RawMessage into a map.
			contextData, ok := contextRaw.(map[string]any)
			if !ok {
				log.Errorf(r.Context(), fmt.Errorf("%s field is not a map", contextKey), "")
				http.Error(w, fmt.Sprintf("%s field is not a map.", contextKey), http.StatusBadRequest)
				return
			}
			var subID any
			switch cfg.Role {
			case "bap":
				subID = contextData["bap_id"]
			case "bpp":
				subID = contextData["bpp_id"]
			}
			ctx := context.WithValue(r.Context(), "subscriber_id", subID)
			// Update keys with UUIDs.
			for _, key := range cfg.UUIDKeys {
				value := uuid.NewString()
				update(contextData, key, value)
				ctx = context.WithValue(ctx, key, value)
			}

			// Marshal the updated JSON.
			updatedBody, err := json.Marshal(data)
			if err != nil {
				http.Error(w, "Failed to marshal updated JSON", http.StatusInternalServerError)
				return
			}

			// Set the updated body.
			r.Body = io.NopCloser(bytes.NewBuffer(updatedBody))
			r.ContentLength = int64(len(updatedBody))
			r = r.WithContext(ctx)

			// Serve the request.
			next.ServeHTTP(w, r)
		})
	}, nil
}

func update(wrapper map[string]any, name string, value any) any {
	field := wrapper[name]
	if field != nil {
		return field
	}
	wrapper[name] = value
	return value
}

func validateConfig(cfg *Config) error {
	if cfg == nil {
		return errors.New("config cannot be nil")
	}
	for _, key := range cfg.UUIDKeys {
		if key == "" {
			return errors.New("UUIDKeys cannot contain empty strings")
		}
	}
	return nil
}
