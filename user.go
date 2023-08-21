package app

import (
	"context"
	"fmt"
	"time"
)

type User struct {
	ID   int
	Name string `json:"name"`

	// Timestamps for user creation & last update.
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt"`
}

func (u User) Validate() error {
	if len(u.Name) == 0 {
		return fmt.Errorf("%s", "missing name")
	}

	return nil
}

type UserService interface {
	CreateUser(ctx context.Context, user *User) error
	FindUserByID(ctx context.Context, userId int) (*User, error)
}

// UserFilter represents a filter passed to FindUsers().
type UserFilter struct {
	// Filtering fields.
	ID *int `json:"id"`

	// Restrict to subset of results.
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}
