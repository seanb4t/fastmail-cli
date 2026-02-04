package cli

// Exit codes for CLI commands
const (
	ExitSuccess      = 0 // Successful operation
	ExitNoToken      = 1 // No token stored
	ExitInvalidToken = 2 // Token expired or revoked
	ExitNetworkError = 3 // Cannot reach API
)
