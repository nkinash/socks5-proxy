package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/nkinash/socks5-proxy/internal/infra"
	"github.com/nkinash/socks5-proxy/internal/service"
)

func main() {
	addr := flag.String("addr", ":1080", "адрес для прослушивания")
	user := flag.String("user", "", "логин (если задан, требует -pass)")
	pass := flag.String("pass", "", "пароль")
	flag.Parse()

	proxy := &service.Proxy{Dialer: &net.Dialer{}}
	if *user != "" && *pass != "" {
		proxy.Authenticator = service.NewStaticAuth(map[string]string{*user: *pass})
	}
	srv := &infra.Server{Proxy: proxy}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := srv.ListenAndServe(ctx, *addr); err != nil {
		log.Fatal(err)
	}
}
