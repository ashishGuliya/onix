package response

import (
	"context"
	"encoding/json"
	"fmt"
)

// Define ErrorType as an enum
type ErrorType string

const (
	SchemaValidationErrorType ErrorType = "SCHEMA_VALIDATION_ERROR"
	InvalidRequestErrorType   ErrorType = "INVALID_REQUEST"
)

type BecknRequest struct {
	Context map[string]interface{} `json:"context,omitempty"`
}

type Error struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Paths   string `json:"paths,omitempty"`
}

type Message struct {
	Ack struct {
		Status string `json:"status,omitempty"`
	} `json:"ack,omitempty"`
	Error *Error `json:"error,omitempty"`
}

type BecknResponse struct {
	Context map[string]interface{} `json:"context,omitempty"`
	Message Message                `json:"message,omitempty"`
}

type ClientFailureBecknResponse struct {
	Context map[string]interface{} `json:"context,omitempty"`
	Error   *Error                 `json:"error,omitempty"`
}

// Map of ErrorType to Error
var errorMap = map[ErrorType]Error{
	SchemaValidationErrorType: {
		Code:    "400",
		Message: "Schema validation failed",
	},
	InvalidRequestErrorType: {
		Code:    "401",
		Message: "Invalid request format",
	},
}

var DefaultError = Error{
	Code:    "500",
	Message: "Internal server error",
}

func Nack(ctx context.Context, tp ErrorType, paths string, body []byte) ([]byte, error) {
	var req BecknRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}

	fmt.Printf("req: %v\n", req)

	errorObj, ok := errorMap[tp]
	if paths != "" {
		errorObj.Paths = paths
	}
	var response BecknResponse

	if !ok {
		response = BecknResponse{
			Context: req.Context,
			Message: Message{
				Ack: struct {
					Status string `json:"status,omitempty"`
				}{
					Status: "NACK",
				},
				Error: &DefaultError,
			},
		}
	} else {
		response = BecknResponse{
			Context: req.Context,
			Message: Message{
				Ack: struct {
					Status string `json:"status,omitempty"`
				}{
					Status: "NACK",
				},
				Error: &errorObj,
			},
		}
	}
	responseObj, err := json.Marshal(response)
	if err != nil {
		fmt.Printf("Parsed response: %+v\n", responseObj)
	}

	return json.Marshal(response)
}

func Acknowledge(ctx context.Context, body []byte) ([]byte, error) {
	var req BecknRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}

	// Create a successful response without an error
	response := BecknResponse{
		Context: req.Context,
		Message: Message{
			Ack: struct {
				Status string `json:"status,omitempty"`
			}{
				Status: "ACK",
			},
		},
	}

	return json.Marshal(response)
}

func handleClientFailure(ctx context.Context, tp ErrorType, body []byte) ([]byte, error) {
	var req BecknRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}

	errorObj, ok := errorMap[tp]
	var response ClientFailureBecknResponse

	if !ok {
		response = ClientFailureBecknResponse{
			Context: req.Context,
			Error:   &DefaultError, // Use the pre-defined DefaultError
		}
	} else {
		response = ClientFailureBecknResponse{
			Context: req.Context,
			Error:   &errorObj,
		}
	}

	return json.Marshal(response)
}
