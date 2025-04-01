package storage

// URLExistsError represents an error when a URL already exists
type URLExistsError struct {
	Message string
}

func (e *URLExistsError) Error() string {
	return e.Message
}

// Is implements the errors.Is interface for compatibility with other errors
func (e *URLExistsError) Is(target error) bool {
	// Check by error message for backward compatibility
	if targetErr, ok := target.(*URLExistsError); ok {
		return e.Message == targetErr.Message
	}
	// For compatibility with simple errors, also check the message
	return target.Error() == e.Message
}

// ErrURLExists is a singleton error for URL already exists
var ErrURLExists = &URLExistsError{Message: "url already exists"}
