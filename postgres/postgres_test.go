package postgres_test

import (
	"fmt"
	"testing"

	"github.com/truescotian/golang-rest-projstructure/postgres"
)

// MustOpenDB returns a new, open DB. Fatal on error.
func MustOpenDB(tb testing.TB) *postgres.DB {
	tb.Helper()

	dsn :=
		fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s search_path=%s", "localhost", "5432", "gregmiller", "go-http-boilerplate", "some_pw", "disable", "public")

	db := postgres.NewDB(dsn)
	if err := db.Open(); err != nil {
		tb.Fatal(err)
	}
	return db
}

// MustCloseDB closes the DB. Fatal on error.
func MustCloseDB(tb testing.TB, db *postgres.DB) {
	tb.Helper()
	if err := db.Close(); err != nil {
		tb.Fatal(err)
	}
}
