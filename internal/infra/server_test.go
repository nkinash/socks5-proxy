package infra_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"golang.org/x/net/proxy"

	"github.com/nkinash/socks5-proxy/internal/infra"
	"github.com/nkinash/socks5-proxy/internal/service"
)

func TestProxyIntegration(t *testing.T) {
	targetMux := http.NewServeMux()
	targetMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "hello from %s", r.URL.Path)
	})

	targetListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer targetListener.Close()

	targetSrv := &http.Server{Handler: targetMux}
	go targetSrv.Serve(targetListener)
	defer targetSrv.Close()

	srv := &infra.Server{Proxy: &service.Proxy{Dialer: &net.Dialer{}}}
	proxyListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go srv.Serve(ctx, proxyListener)

	proxyAddr := proxyListener.Addr().String()

	socksDialer, err := proxy.SOCKS5("tcp", proxyAddr, nil, proxy.Direct)
	if err != nil {
		t.Fatalf("create socks5 dialer: %v", err)
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			Dial:                  socksDialer.Dial,
			ResponseHeaderTimeout: 5 * time.Second,
		},
		Timeout: 10 * time.Second,
	}

	targetURL := fmt.Sprintf("http://%s/hello", targetListener.Addr().String())
	resp, err := httpClient.Get(targetURL)
	if err != nil {
		t.Fatalf("http get через прокси: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("ожидался 200, получен %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	expected := "hello from /hello"
	if string(body) != expected {
		t.Fatalf("ожидалось %q, получено %q", expected, string(body))
	}
}
