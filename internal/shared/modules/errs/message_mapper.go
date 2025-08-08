package errs

var mapping = map[string]string{
	// Authentication & Authorization
	codeUnauthorized: "Unauthorized request",
	codeForbidden:    "Access forbidden",
	codeInvalidToken: "Invalid authentication token",
	codeTokenExpired: "Authentication token has expired",

	// Input & Validation
	codeInvalidArgument: "Invalid input provided",

	// Resource Status
	codeNotFound:      "Resource not found",
	codeAlreadyExists: "Resource already exists",
	codeConflict:      "Resource state conflict",
	codeGone:          "Resource is no longer available",

	// Business Logic
	codeEmailInUse:  "Email address is already in use",
	codeRateLimited: "Too many requests, please try again later",

	// External Services
	codeExternalService: "External service error",
	codeDatabaseError:   "Database operation failed",
	codeNetworkError:    "Network communication error",

	// Server Errors
	codeInternalError:      "Internal server error",
	codeNotImplemented:     "Feature not implemented",
	codeServiceUnavailable: "Service is currently unavailable",
	codeTimeout:            "Request timed out",

	// Data
	codeInvalidState: "Invalid data state",

	// Unknown
	codeUnknown: "Unknown error",
}

func mapMessage(code string) string {
	if msg, ok := mapping[code]; ok {
		return msg
	}
	return "Unknown error"
}
