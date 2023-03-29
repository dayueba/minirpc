package minirpc

import "time"

type ServerOptions struct {
	addr    string
	timeout time.Duration
}

type ServerOption func(*ServerOptions)

func WithTimeout(timeout time.Duration) ServerOption {
	return func(o *ServerOptions) { o.timeout = timeout }
}
