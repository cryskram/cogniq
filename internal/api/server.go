package api

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"

	"github.com/cryskram/relith/internal/config"
	"github.com/cryskram/relith/internal/db"
	"github.com/cryskram/relith/internal/indexer"
	"github.com/cryskram/relith/internal/search"
)

type Server struct {
	http   *http.Server
	listen net.Listener
	logger zerolog.Logger
	cfg    config.DaemonConfig
}

func New(database *sql.DB, logger zerolog.Logger, cfg *config.Config) *Server {
	h := &handlers{
		queries:  db.New(database),
		indexer:  indexer.New(database, logger, cfg.Indexer),
		searcher: search.New(database, logger, cfg.Search),
	}

	mux := http.NewServeMux()
	mux.Handle("GET /", dashboardHandler())
	mux.HandleFunc("GET /v1/health", h.health)
	mux.HandleFunc("GET /v1/repos", h.listRepos)
	mux.HandleFunc("POST /v1/repos", h.createRepo)
	mux.HandleFunc("GET /v1/repos/{id}", h.getRepo)
	mux.HandleFunc("DELETE /v1/repos/{id}", h.deleteRepo)
	mux.HandleFunc("POST /v1/repos/{id}/index", h.indexRepo)
	mux.HandleFunc("GET /v1/search", h.search)
	mux.HandleFunc("GET /v1/content", h.content)

	httpSrv := &http.Server{
		Handler:     withLogging(mux, logger),
		ReadTimeout: 10 * time.Second,
		IdleTimeout: 60 * time.Second,
	}

	return &Server{
		http:   httpSrv,
		logger: logger.With().Str("component", "api").Logger(),
		cfg:    cfg.Daemon,
	}
}

func (s *Server) Start() error {
	socketPath := s.cfg.Socket

	if val, ok := os.LookupEnv("RELITH_DAEMON_SOCKET"); ok && val == "" {
		socketPath = ""
	}

	if socketPath != "" {
		os.Remove(socketPath)
		listener, err := net.Listen("unix", socketPath)
		if err != nil {
			return fmt.Errorf("unix listen: %w", err)
		}
		if err := os.Chmod(socketPath, 0660); err != nil {
			listener.Close()
			return fmt.Errorf("socket chmod: %w", err)
		}
		s.listen = listener
		s.logger.Info().Str("socket", socketPath).Msg("listening on unix socket")
	} else {
		addr := fmt.Sprintf("%s:%d", s.cfg.TCPHost, s.cfg.TCPPort)
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("tcp listen: %w", err)
		}
		s.listen = listener
		s.logger.Info().Str("addr", addr).Msg("listening on tcp")
	}

	go func() {
		if err := s.http.Serve(s.listen); err != nil && err != http.ErrServerClosed {
			s.logger.Error().Err(err).Msg("server error")
		}
	}()

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := s.http.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	if s.cfg.Socket != "" {
		os.Remove(s.cfg.Socket)
	}

	s.logger.Info().Msg("server stopped")
	return nil
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func withLogging(next http.Handler, logger zerolog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		logger.Debug().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", rw.status).
			Dur("duration", time.Since(start)).
			Msg("request")
	})
}
