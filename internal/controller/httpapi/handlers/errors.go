package handlers

import "errors"

// Group of errors for the utils.
var (
	// ErrTrustedSubnetEmpty is the error for when the trusted subnet is empty.
	ErrTrustedSubnetEmpty = errors.New("trusted subnet is empty")
	// ErrClientIPEmpty is the error for when the client IP is empty.
	ErrClientIPEmpty = errors.New("client IP is empty")
	// ErrInvalidTrustedSubnetCIDR is the error for when the trusted subnet CIDR is invalid.
	ErrInvalidTrustedSubnetCIDR = errors.New("invalid trusted subnet CIDR")
	// ErrInvalidClientIP is the error for when the client IP is invalid.
	ErrInvalidClientIP = errors.New("invalid client IP")
)
