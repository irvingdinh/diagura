package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	nethttp "net/http"

	"go.uber.org/fx"

	"localhost/app/core/config"
	"localhost/app/core/config/rule"
)

type RouteRegistrar interface {
	RegisterRoutes(mux *nethttp.ServeMux)
}

type Server interface{}

func NewServer(params struct {
	fx.In
	Lifecycle  fx.Lifecycle
	Registrars []RouteRegistrar `group:"routes"`
}) Server {
	config.SetDefaults(config.Values{
		"host":               "127.0.0.1",
		"port":               48310,
		"http.read_timeout":  "5s",
		"http.write_timeout": "10s",
		"http.idle_timeout":  "120s",
	})
	config.SetRule("host", rule.Required)
	config.SetRule("port", rule.Required, rule.Between(1, 65535))
	config.SetRule("http.read_timeout", rule.Required, rule.Duration)
	config.SetRule("http.write_timeout", rule.Required, rule.Duration)
	config.SetRule("http.idle_timeout", rule.Required, rule.Duration)

	s := &serverImpl{registrars: params.Registrars}
	params.Lifecycle.Append(fx.Hook{
		OnStart: s.start,
		OnStop:  s.stop,
	})
	return s
}

type serverImpl struct {
	registrars []RouteRegistrar
	server     *nethttp.Server
}

func (s *serverImpl) start(_ context.Context) error {
	mux := nethttp.NewServeMux()
	for _, r := range s.registrars {
		r.RegisterRoutes(mux)
	}
	mux.HandleFunc("GET /api", handleAPI)

	addr := fmt.Sprintf("%s:%d", config.GetString("host"), config.GetInt("port"))
	s.server = &nethttp.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  config.GetDuration("http.read_timeout"),
		WriteTimeout: config.GetDuration("http.write_timeout"),
		IdleTimeout:  config.GetDuration("http.idle_timeout"),
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	go func() { _ = s.server.Serve(ln) }()

	return nil
}

func (s *serverImpl) stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func handleAPI(w nethttp.ResponseWriter, _ *nethttp.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "OK"})
}
