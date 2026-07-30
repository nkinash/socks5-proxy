package infra

import (
	"context"
	"log"
	"net"

	"github.com/nkinash/socks5-proxy/internal/service"
)

type Server struct {
	Proxy *service.Proxy
}

func (s *Server) ListenAndServe(ctx context.Context, addr string) error {
	lc := net.ListenConfig{}
	l, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	log.Printf("socks5 listening on %s", addr)
	return s.Serve(ctx, l)
}

func (s *Server) Serve(ctx context.Context, l net.Listener) error {
	defer l.Close()

	go func() {
		<-ctx.Done()
		l.Close()
	}()

	for {
		conn, err := l.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				log.Printf("accept: %v", err)
				continue
			}
		}
		go s.Proxy.Handle(ctx, conn)
	}
}
