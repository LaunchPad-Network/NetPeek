package traceroute

import "errors"

var (
	ErrEmptyTarget  = errors.New("empty target for traceroute")
	ErrNotSupported = errors.New("traceroute binary not configured")
)
