package server

import (
	"fmt"
    "log/slog"
	"net/http"
)

type Server struct {
	CompletedConfig

	Handler http.Handler
	Log     *slog.Logger
}

// Only a preparedServer can be Run
type preparedServer struct {
	*Server
}

func New(c CompletedConfig, handler http.Handler, log *slog.Logger) (*Server, error) {
	return &Server{
		CompletedConfig: c,
		Handler:         handler,
		Log:             log,
	}, nil
}

func (s *Server) PrepareRun() preparedServer {
	return preparedServer{s}
}

func (s preparedServer) Run() error {
	s.Log.Info(fmt.Sprintf("Listening on address %s", s.Options.Address))

	if s.SecureServing {
		return http.ListenAndServeTLS(s.Options.Address, s.Options.CertFile, s.Options.KeyFile, s.Handler)
	}

	return http.ListenAndServe(s.Options.Address, s.Handler)
}
