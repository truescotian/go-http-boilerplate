package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	app "github.com/truescotian/golang-rest-projstructure"
)

// ShutdownTimeout is the time given for outstanding requests to finish before shutdown.
const ShutdownTimeout = 1 * time.Second

// Server is an HTTP server. It wraps all HTTP functionality used by the
// application so that dependent packages (such as cmd/wtfd) do not need
// to reference the "net/http" package at all.
type Server struct {
	ln     net.Listener
	server *http.Server
	router *mux.Router

	// Bind address & domain for the server's listener.
	Addr string

	// Services used by the various HTTP routes.
	UserService app.UserService
}

func NewServer() *Server {
	s := &Server{
		server: &http.Server{},
		router: mux.NewRouter(),
	}

	// Report panics to external service
	s.router.Use(reportPanic)

	// Router is wrapped by another function handler to perform some
	// middleware-like tasks that cannot be performed by actual middleware.
	s.server.Handler = http.HandlerFunc(s.serveHTTP)

	// Setup error handling routes
	s.router.NotFoundHandler = http.HandlerFunc(s.handleNotFound)

	{
		r := s.router.PathPrefix("/").Subrouter()
		s.registerUserRoutes(r)
	}

	return s

}

func (s *Server) serveHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// handleNotFound handles requests to routes that don't exist.
func (s *Server) handleNotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	return
}

// reportPanic is middleware for catching panics and reporting them
func reportPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				app.ReportPanic(err)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// URL returns the local base URL of the running server.
func (s *Server) URL() string {
	return fmt.Sprintf("%s://%s:%d", "http", "localhost", 6060)
}

// Open begins listening on the bind address
func (s *Server) Open() (err error) {
	if s.ln, err = net.Listen("tcp", s.Addr); err != nil {
		return err
	}

	// Begin serving requests on the listener. We use Serve() instead of
	// ListenAndServe() because it allows us to check for listen errors (such
	// as trying to use an already open port) synchronously.
	go s.server.Serve(s.ln)

	return nil
}

// Close gracefully shuts down the server.
func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()
	return s.server.Shutdown(ctx)
}

// ListenAndServeDebug runs an HTTP server with /debug endpoints (e.g. pprof, vars).
func ListenAndServeDebug() error {
	h := http.NewServeMux()
	// TODO
	// h.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(":6060", h)
}
