package errs

const (
	// Authentication & Authorization.
	codeUnauthorized = "UNAUTHORIZED" // Not authenticated
	codeForbidden    = "FORBIDDEN"    // Not authorized
	codeInvalidToken = "INVALID_TOKEN"
	codeTokenExpired = "TOKEN_EXPIRED"

	// Input & Validation.
	codeInvalidArgument = "INVALID_ARGUMENT" // Generic validation codeor

	// Resource Status.
	codeNotFound      = "NOT_FOUND"      // Resource doesn't exist
	codeAlreadyExists = "ALREADY_EXISTS" // Duplicate resource
	codeConflict      = "CONFLICT"       // Resource state conflict
	codeGone          = "GONE"           // Resource no longer available

	// Business Logic.
	codeEmailInUse  = "EMAIL_IN_USE"
	codeRateLimited = "RATE_LIMITED"

	// External Services.
	codeExternalService = "EXTERNAL_SERVICE_ERROR"
	codeDatabaseError   = "DATABASE_ERROR"
	codeNetworkError    = "NETWORK_ERROR"

	// Server codeors.
	codeInternalError      = "INTERNAL_ERROR" // Generic server codeor
	codeNotImplemented     = "NOT_IMPLEMENTED"
	codeServiceUnavailable = "SERVICE_UNAVAILABLE"
	codeTimeout            = "TIMEOUT"

	// Data.
	codeInvalidState = "INVALID_STATE"

	// Bad Request.
	codeBadRequest = "BAD_REQUEST"

	// Unknown.
	codeUnknown = "UNKNOWN"
)
