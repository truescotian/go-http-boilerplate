package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	app "github.com/truescotian/golang-rest-projstructure"
	"github.com/truescotian/golang-rest-projstructure/http"
	"github.com/truescotian/golang-rest-projstructure/postgres"
)

func main() {
	// Setup signal handlers
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() { <-c; cancel() }()

	// This type lets us share setup code with end-to-end tests.
	m := NewMain()

	if err := m.Run(ctx); err != nil {
		m.Close()
		fmt.Fprintln(os.Stderr, err)
		app.ReportError(ctx, err)
		os.Exit(1)
	}

	// wait for CTRL-C
	<-ctx.Done()

	// clean up
	if err := m.Close(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type Main struct {
	DB *postgres.DB

	// HTTP server for handling HTTP communication.
	// SQLite services are attached to it before running.
	HTTPServer *http.Server

	// Services exposed for end-to-end tests.
	UserService app.UserService
}

func NewMain() *Main {
	connectionString :=
		fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s search_path=%s", "localhost", "5432", "gregmiller", "go-http-boilerplate", "some_pw", "disable", "public")

	return &Main{
		DB:         postgres.NewDB(connectionString),
		HTTPServer: http.NewServer(),
	}
}

func (m *Main) Close() error {
	if m.HTTPServer != nil {
		if err := m.HTTPServer.Close(); err != nil {
			return err
		}
	}
	if m.DB != nil {
		if err := m.DB.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (m *Main) Run(ctx context.Context) (err error) {
	if err := m.DB.Open(); err != nil {
		return fmt.Errorf("cannot open db: %w", err)
	}

	// Initialize postgres-backed services.
	userService := postgres.NewUserService(m.DB)

	// Attach user service to Main for testing.
	m.UserService = userService

	m.HTTPServer.Addr = ":6060"
	m.HTTPServer.UserService = userService

	if err := m.HTTPServer.Open(); err != nil {
		return err
	}

	// Enable internal debug endpoints.
	go func() { http.ListenAndServeDebug() }()

	log.Printf("running: url=%q debug=http://localhost:6060 dsn=%q", m.HTTPServer.URL(), m.DB.DSN)

	return nil
}
