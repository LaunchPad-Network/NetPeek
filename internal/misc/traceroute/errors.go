package traceroute

import "errors"

var (
	ErrEmptyTarget   = errors.New("empty target for traceroute")
	ErrNotSupported  = errors.New("traceroute not supported")
	ErrInvalidTarget = errors.New("invalid target")
)
