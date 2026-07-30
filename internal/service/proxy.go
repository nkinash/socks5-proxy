package service

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"slices"

	"github.com/nkinash/socks5-proxy/internal/domain"
	"github.com/nkinash/socks5-proxy/internal/protocol"
)

type Proxy struct {
	Dialer        domain.Dialer
	Authenticator domain.Authenticator
}

func (p *Proxy) Handle(ctx context.Context, client net.Conn) {
	defer client.Close()

	methods, err := protocol.ReadHandshake(client)
	if err != nil {
		log.Printf("handshake read: %v", err)
		return
	}

	if p.Authenticator != nil {
		if !slices.Contains(methods, domain.AuthUserPass) {
			protocol.WriteHandshake(client, domain.AuthNoAccep)
			log.Printf("handshake: no user/pass in client methods")
			return
		}
		if err := protocol.WriteHandshake(client, domain.AuthUserPass); err != nil {
			log.Printf("handshake write: %v", err)
			return
		}
		user, pass, err := protocol.ReadUserPass(client)
		if err != nil {
			protocol.WriteUserPassReply(client, false)
			log.Printf("userpass read: %v", err)
			return
		}
		if !p.Authenticator.Authenticate(user, pass) {
			protocol.WriteUserPassReply(client, false)
			log.Printf("auth: bad credentials for %q", user)
			return
		}
		if err := protocol.WriteUserPassReply(client, true); err != nil {
			log.Printf("userpass reply write: %v", err)
			return
		}
	} else {
		if !slices.Contains(methods, domain.AuthNone) {
			protocol.WriteHandshake(client, domain.AuthNoAccep)
			log.Printf("handshake: no no-auth in client methods")
			return
		}
		if err := protocol.WriteHandshake(client, domain.AuthNone); err != nil {
			log.Printf("handshake write: %v", err)
			return
		}
	}

	req, err := protocol.ReadRequest(client)
	if err != nil {
		log.Printf("request read: %v", err)
		protocol.WriteReply(client, domain.RepGeneralFailure)
		return
	}

	if req.Cmd != domain.CmdConnect {
		protocol.WriteReply(client, domain.RepCmdNotSupported)
		log.Printf("unsupported cmd: %d", req.Cmd)
		return
	}

	targetAddr := fmt.Sprintf("%s:%d", req.Addr.Host, req.Addr.Port)
	target, err := p.Dialer.DialContext(ctx, "tcp", targetAddr)
	if err != nil {
		log.Printf("dial target %s: %v", targetAddr, err)
		protocol.WriteReply(client, domain.RepHostUnreachable)
		return
	}
	defer target.Close()

	if err := protocol.WriteReply(client, domain.RepSuccess); err != nil {
		log.Printf("reply write: %v", err)
		return
	}

	log.Printf("connected %s → %s", client.RemoteAddr(), targetAddr)
	relay(client, target)
}

func relay(client, target net.Conn) {
	done := make(chan struct{}, 2)

	go func() {
		io.Copy(target, client)
		done <- struct{}{}
	}()
	go func() {
		io.Copy(client, target)
		done <- struct{}{}
	}()

	<-done
	log.Printf("closed %s", client.RemoteAddr())
}
