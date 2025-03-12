package reqpreprocessor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"onix/shared/log"

	"github.com/google/uuid"
)



// updateFieldsMiddleware updates fields based on relative paths in the config.
func New() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Read the request body.

			var data map[string]any
			if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
				log.Errorf(r.Context(), err, "json.NewDecoder(r.Body): %v", err)
				http.Error(w, "Failed to decode request body", http.StatusBadRequest)
				return
			}
			// Get the "context" field as a RawMessage.
			contextRaw := data["context"]
			if contextRaw == nil {
				log.Errorf(r.Context(), fmt.Errorf("context field not found"), "")
				http.Error(w, "Context field not found.", http.StatusBadRequest)
				return
			}
			// Unmarshal the "context" RawMessage into a map.
			contextData, ok := contextRaw.(map[string]any)
			if !ok {
				log.Errorf(r.Context(), fmt.Errorf("context field not found"), "")
				http.Error(w, "Context field not found.", http.StatusBadRequest)
				return
			}

			// Update the TransactionID in the Context.
			txnID := update(contextData, "transaction_id", uuid.NewString())
			msgID := update(contextData, "message_id", uuid.NewString())
			ctx := context.WithValue(r.Context(), "transaction_id", txnID)
			ctx = context.WithValue(ctx, "message_id", msgID)

			// Marshal the updated JSON.
			updatedBody, err := json.Marshal(data)
			if err != nil {
				http.Error(w, "Failed to marshal updated JSON", http.StatusInternalServerError)
				return
			}

			// Set the updated body.
			r.Body = io.NopCloser(bytes.NewBuffer(updatedBody))
			r.ContentLength = int64(len(updatedBody))

			// Serve the requestr
		})
	}
}

func update(wrapper map[string]any, name string, value any) any {
	field := wrapper[name]
	if field != nil {
		return field
	}
	wrapper[name] = value
	return value
}
