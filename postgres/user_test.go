package postgres_test

import (
	"context"
	"reflect"
	"testing"

	app "github.com/truescotian/golang-rest-projstructure"
	"github.com/truescotian/golang-rest-projstructure/postgres"
)

func TestUserService_CreateUser(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		p := postgres.NewUserService(db)

		u := &app.User{
			Name: "greg",
		}

		if err := p.CreateUser(context.Background(), u); err != nil {
			t.Fatal(err)
		} else if got, want := u.ID, 1; got != want {
			t.Fatalf("ID=%v, want %v", got, want)
		} else if u.CreatedAt.IsZero() {
			t.Fatal("expected created at")
		} else if u.UpdatedAt.IsZero() {
			t.Fatal("expected updated at")
		}

		// Create second user with email.
		u2 := &app.User{Name: "jane"}
		if err := p.CreateUser(context.Background(), u2); err != nil {
			t.Fatal(err)
		} else if got, want := u2.ID, 2; got != want {
			t.Fatalf("ID=%v, want %v", got, want)
		}

		// Fetch user from database & compare.
		if other, err := p.FindUserByID(context.Background(), 1); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(u, other) {
			t.Fatalf("mismatch: %#v != %#v", u, other)
		}
	})
}
