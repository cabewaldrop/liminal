package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/cabewaldrop/liminal/pkg/middleware/metrics"
	"github.com/cabewaldrop/liminal/pkg/middleware/ratelimit"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, World!"))
}

func NewRouter() *mux.Router {
	m := mux.NewRouter()

	rl := ratelimit.NewRateLimiter("IP", http.HandlerFunc(HelloHandler))
	reg := prometheus.NewRegistry()
	metrics := metrics.NewMetrics(rl, reg)
	m.Handle("/", metrics)
	m.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
	return m
}

func NewServer() *http.Server {
	r := NewRouter()
	return &http.Server{
		Addr:         "127.0.0.1:8000",
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	server := NewServer()

	defer cancel()
	go func() {
		server.ListenAndServe()
	}()
	log.Info().Msgf("Server listening on: %s", server.Addr)

	<-ctx.Done()

	ctx, done := context.WithTimeout(context.Background(), 30*time.Second)
	defer done()
	server.Shutdown(ctx)
}
