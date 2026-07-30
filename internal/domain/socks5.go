package domain

import (
	"context"
	"net"
)

type Addr struct {
	Type byte
	Host string
	Port uint16
}

type Request struct {
	Cmd  byte
	Addr Addr
}

type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

type Authenticator interface {
	Authenticate(user, pass string) bool
}
