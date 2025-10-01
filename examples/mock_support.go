package main

import (
	"context"
	"net"
	"net/http"
	"time"
)

// startMockOKServer starts a lightweight HTTP server that returns 200 OK on /ok and /
// It listens on 127.0.0.1 with an ephemeral port and returns the base URL and a stop function.
func startMockOKServer() (baseURL string, stop func()) {
	mux := http.NewServeMux()
    handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}
    // failure endpoint returns 500 to simulate branch failure
    fail := func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusInternalServerError)
        _, _ = w.Write([]byte("FAIL"))
    }
	mux.HandleFunc("/ok", handler)
    mux.HandleFunc("/fail", fail)
	mux.HandleFunc("/", handler)

	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	srv := &http.Server{Handler: mux}
	go func() { _ = srv.Serve(ln) }()

	stop = func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}
	return "http://" + ln.Addr().String(), stop
}
