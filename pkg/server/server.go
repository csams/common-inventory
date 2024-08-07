package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type Server struct {
	SecureServing bool

	ServingCertFile string
	PrivateKeyFile  string

	HttpServer *http.Server
	Log        *slog.Logger
}

type preparedServer struct {
	*Server
}

func New(c CompletedConfig, handler http.Handler, log *slog.Logger) *Server {
	return &Server{
		HttpServer: &http.Server{
			Addr:         c.Options.Addr,
			ReadTimeout:  time.Duration(c.Options.ReadTimeout) * time.Second,
			WriteTimeout: time.Duration(c.Options.WriteTimeout) * time.Second,
			TLSConfig:    c.TLSConfig,
			Handler:      handler,
		},
		Log: log,
	}
}

// Only a preparedServer can be Run, so we can't start an incorrectly configured server
func (s *Server) Run() error {
	s.Log.Info(fmt.Sprintf("Listening on address %s", s.HttpServer.Addr))

	if s.SecureServing {
		return s.HttpServer.ListenAndServeTLS(s.ServingCertFile, s.PrivateKeyFile)
	}

	return s.HttpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.HttpServer.Shutdown(ctx)
}
