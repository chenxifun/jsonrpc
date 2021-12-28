package config

import "time"

const defPort = 8003

func DefaultConfig() Config {
	return Config{
		Origins:   []string{"*"},
		Vhosts:    []string{"*"},
		Cors:      []string{"*"},
		Hosts:     "localhost",
		Port:      defPort,
		EnableRPC: true,
		EnableWS:  true,
	}
}

type Config struct {
	Vhosts  []string
	Cors    []string
	Origins []string

	Hosts string
	Port  int

	EnableWS  bool
	EnableRPC bool

	// ReadTimeout is the maximum duration for reading the entire
	// request, including the body.
	//
	// Because ReadTimeout does not let Handlers make per-request
	// decisions on each request body's acceptable deadline or
	// upload rate, most users will prefer to use
	// ReadHeaderTimeout. It is valid to use them both.
	ReadTimeout time.Duration

	// WriteTimeout is the maximum duration before timing out
	// writes of the response. It is reset whenever a new
	// request's header is read. Like ReadTimeout, it does not
	// let Handlers make decisions on a per-request basis.
	WriteTimeout time.Duration

	// IdleTimeout is the maximum amount of time to wait for the
	// next request when keep-alives are enabled. If IdleTimeout
	// is zero, the value of ReadTimeout is used. If both are
	// zero, ReadHeaderTimeout is used.
	IdleTimeout time.Duration
}
