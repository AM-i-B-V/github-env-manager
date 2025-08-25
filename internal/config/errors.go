package config

import "errors"

var (
	// ErrInvalidPort is returned when the port configuration is invalid
	ErrInvalidPort = errors.New("invalid port number")

	// ErrInvalidHost is returned when the host configuration is invalid
	ErrInvalidHost = errors.New("invalid host configuration")
)
