package response

import (
	"context"
	"encoding/json"
)

// Error represents a custom error with code, message, and description.
type error struct {
    Code int `json:"code"`
    Message string `json:"message"`
    Description string `json:"description"`
}
// :Todo Change to enum


type ErrorType string
// Error Keys as Constants


const (

	SchemaValidationErrorType ErrorType ="SchemaValidationErrorType"

)
const (
    SchemaValidationErrorKey = "ValidationError"
    RequestTimedoutErrorKey = "RequestTimedoutError"
    UnauthenticatedErrorKey= "UnauthorizedError"
    InternalServerErrorKey = "InternalServerError"
)

// Error Map
// Will need to update the Error map with all the errors in core beckn spec.
// Ravi Prakash to provide the final codes and error message
var errorMap = map[ErrorType] Error {
    SchemaValidationErrorKey: {
        Code: 400,
        Message: "Validation failed",
        Description: "The provided input data is invalid.",
    },
    RequestTimedoutErrorKey: {
        Code: 500,
        Message: "Request has been timed out.",
        Description: "Request has been timed out.",
    },
	UnauthenticatedErrorKey: {
        Code: 401,
        Message: "Unauthorized",
        Description: "Authentication failed.",
    },
    InternalServerErrorKey: {
        Code: 500,
        Message: "Internal Server Error",
        Description: "An unexpected error occurred on the server.",
    },
}


// var DefaultError = errorMap[InternalServerErrorKey]
// // BecknRequestContext as interface{} (any)
// type BecknRequestContext interface {}

// // Message struct defined outside BecknResponse
// type Message struct {
//     Ack struct {
//         Status string `json:"status"`
//     }
//     `json:"ack"`
//     Error * ErrorObject `json:"error,omitempty"`
// }



// // BecknResponse represents the Beckn response structure.
// type BecknResponse struct {
//     Context interface {}
//     `json:"context"`
//     Message Message `json:"message"`
// }


// // acknowledge mirrors your Node.js acknowledge function.
// func acknowledge(context BecknRequestContext)([] byte, error) {
//     var response BecknResponse

//     response = BecknResponse {
//         Context: context,
//         Message: Message {
//             Ack: struct {
//                 Status string `json:"status"`
//             } {
//                 Status: "ACK",
//             },
//         },
//     }func getNackResponse(errorKey string, context BecknRequestContext)([] byte, error) {
// 		errorObj, ok: = errorMap[errorKey]
// 		var response BecknResponse
	
// 		if !ok {
// 			response = BecknResponse {
// 				Context: context,
// 				Message: Message {errorKey string, context BecknRequestContext
// 					},
// 					Error: & DefaultError, // Use the pre-defined DefaultError
// 				},
// 			}
// 		} else {
// 			response = BecknResponse {
// 				Context: context,
// 				Message: Message {
// 					Ack: struct {
// 						Status string `json:"status"`
// 					} {
// 						Status: "NACK",
// 					},
// 					Error: & errorObj,
// 				},
// 			}
// 		}
	
// 		jsonBytes, err: = json.Marshal(response)
// 		if err != nil {
// 			return nil, err
// 		}
// 		return jsonBytes, nil
// 	}
	
// 	 err: = json.Marshal(response)
//     if err != nil {
//         return nil, err
//     }
//     return jsonBytes, nil
// }

type reqWithCtx struct{

	Context context.Context
}

func Nack(ctx context.Context, tp ErrorType, body []byte )([] byte, error){
var req reqWithCtx
if err:= json.Unmarshal(body, &req); err!= nil{
	return nil, err
}
req.Context
    
	errorObj, ok: = errorMap[tp]
    var response BecknResponse

    if !ok {
        response = BecknResponse {
            Context: context,
            Message: Message {
                Ack: struct {
                    Status string `json:"status"`
                } {
                    Status: "NACK",
                },
                Error: & DefaultError, // Use the pre-defined DefaultError
            },
        }
    } else {
        response = BecknResponse {
            Context: context,
            Message: Message {
                Ack: struct {
                    Status string `json:"status"`
                } {
                    Status: "NACK",
                },
                Error: & errorObj,
            },
        }
    }

    return json.Marshal(response)
    
}

