package response

// Error represents a custom error with code, message, and description.
type error struct {
	Code        int    `json:"code"`
	Message     string `json:"message"`
	Description string `json:"description"`
}
