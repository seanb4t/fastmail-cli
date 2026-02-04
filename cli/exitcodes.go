package cli

// Exit codes for CLI commands
const (
	ExitSuccess      = 0 // Successful operation
	ExitNoToken      = 1 // No token stored
	ExitInvalidToken = 2 // Token expired or revoked
	ExitNetworkError = 3 // Cannot reach API
)

// AuthStatusError wraps an error with an exit code for auth status commands.
type AuthStatusError struct {
	Code    int
	Message string
}

// Error implements the error interface.
func (e *AuthStatusError) Error() string {
	return e.Message
}
