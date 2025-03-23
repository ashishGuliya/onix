package response

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/ashishGuliya/onix/pkg/model"
)

type ErrorType string

const (
	SchemaValidationErrorType ErrorType = "SCHEMA_VALIDATION_ERROR"
	InvalidRequestErrorType   ErrorType = "INVALID_REQUEST"
)

type BecknRequest struct {
	Context map[string]interface{} `json:"context,omitempty"`
}

// type Message struct {
// 	Ack struct {
// 		Status string `json:"status,omitempty"`
// 	} `json:"ack,omitempty"`
// 	Error *Error `json:"error,omitempty"`
// }

// type BecknResponse struct {
// 	Context map[string]interface{} `json:"context,omitempty"`
// 	Message Message                `json:"message,omitempty"`
// }

// type ClientFailureBecknResponse struct {
// 	Context map[string]interface{} `json:"context,omitempty"`
// 	Error   *Error                 `json:"error,omitempty"`
// }

// var errorMap = map[ErrorType]Error{
// 	SchemaValidationErrorType: {
// 		Code:    "400",
// 		Message: "Schema validation failed",
// 	},
// 	InvalidRequestErrorType: {
// 		Code:    "401",
// 		Message: "Invalid request format",
// 	},
// }

// var DefaultError = Error{
// 	Code:    "500",
// 	Message: "Internal server error",
// }

// func Nack(ctx context.Context, tp ErrorType, paths string, body []byte) ([]byte, error) {
// 	var req BecknRequest
// 	if err := json.Unmarshal(body, &req); err != nil {
// 		return nil, fmt.Errorf("failed to parse request: %w", err)
// 	}

// 	errorObj, ok := errorMap[tp]
// 	if paths != "" {
// 		errorObj.Paths = paths
// 	}

// 	var response BecknResponse

// 	if !ok {
// 		response = BecknResponse{
// 			Context: req.Context,
// 			Message: Message{
// 				Ack: struct {
// 					Status string `json:"status,omitempty"`
// 				}{
// 					Status: "NACK",
// 				},
// 				Error: &DefaultError,
// 			},
// 		}
// 	} else {
// 		response = BecknResponse{
// 			Context: req.Context,
// 			Message: Message{
// 				Ack: struct {
// 					Status string `json:"status,omitempty"`
// 				}{
// 					Status: "NACK",
// 				},
// 				Error: &errorObj,
// 			},
// 		}
// 	}

// 	return json.Marshal(response)
// }

// func Ack(ctx context.Context, body []byte) ([]byte, error) {
// 	var req BecknRequest
// 	if err := json.Unmarshal(body, &req); err != nil {
// 		return nil, fmt.Errorf("failed to parse request: %w", err)
// 	}

// 	response := BecknResponse{
// 		Context: req.Context,
// 		Message: Message{
// 			Ack: struct {
// 				Status string `json:"status,omitempty"`
// 			}{
// 				Status: "ACK",
// 			},
// 		},
// 	}

// 	return json.Marshal(response)
// }

// func HandleClientFailure(ctx context.Context, tp ErrorType, body []byte) ([]byte, error) {
// 	var req BecknRequest
// 	if err := json.Unmarshal(body, &req); err != nil {
// 		return nil, fmt.Errorf("failed to parse request: %w", err)
// 	}

// 	errorObj, ok := errorMap[tp]
// 	var response ClientFailureBecknResponse

// 	if !ok {
// 		response = ClientFailureBecknResponse{
// 			Context: req.Context,
// 			Error:   &DefaultError,
// 		}
// 	} else {
// 		response = ClientFailureBecknResponse{
// 			Context: req.Context,
// 			Error:   &errorObj,
// 		}
// 	}

// 	return json.Marshal(response)
// }

func SendAck(w http.ResponseWriter) {
	// Create the response object
	resp := &model.Response{
		Message: model.Message{
			Ack: model.Ack{
				Status: model.StatusACK,
			},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "failed to marshal response", http.StatusInternalServerError)
		return
	}

	// Set headers and write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// nack sends a negative acknowledgment (NACK) response with an error message.
func nack(w http.ResponseWriter, err *model.Error, status int) {
	// Create the NACK response object
	resp := &model.Response{
		Message: model.Message{
			Ack: model.Ack{
				Status: model.StatusNACK,
			},
			Error: err,
		},
	}

	// Marshal the response to JSON
	data, jsonErr := json.Marshal(resp)
	if jsonErr != nil {
		http.Error(w, "failed to marshal response", http.StatusInternalServerError)
		return
	}

	// Set headers and write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status) // Assuming NACK means a bad request
	w.Write(data)
}

func internalServerError(ctx context.Context) *model.Error {
	return &model.Error{
		Message: fmt.Sprintf("Internal server error, MessageID: %s", ctx.Value(model.MsgIDKey)),
	}
}

// SendNack sends a negative acknowledgment (NACK) response with an error message.
func SendNack(ctx context.Context, w http.ResponseWriter, err error) {
	var schemaErr *model.SchemaValidationErr
	var signErr *model.SignValidationErr
	var badReqErr *model.BadReqErr
	var notFoundErr *model.NotFoundErr

	switch {
	case errors.As(err, &schemaErr): // Custom application error
		nack(w, schemaErr.BecknError(), http.StatusBadRequest)
		return
	case errors.As(err, &signErr):
		nack(w, signErr.BecknError(), http.StatusUnauthorized)
		return
	case errors.As(err, &badReqErr):
		nack(w, badReqErr.BecknError(), http.StatusBadRequest)
		return
	case errors.As(err, &notFoundErr):
		nack(w, notFoundErr.BecknError(), http.StatusNotFound)
		return
	default:
		nack(w, internalServerError(ctx), http.StatusInternalServerError)
		return
	}
}

func BecknError(ctx context.Context, err error, status int) *model.Error {
	msg := err.Error()
	msgID := ctx.Value(model.MsgIDKey)
	if status == http.StatusInternalServerError {

		msg = "Internal server error"
	}
	return &model.Error{
		Message: fmt.Sprintf("%s. MessageID: %s.", msg, msgID),
		Code:    string(status),
	}
}
