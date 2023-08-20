package app

import (
	"context"
	"fmt"
)

type User struct {
	ID   int
	Name string `json:"name"`
}

func (u User) Validate() error {
	if len(u.Name) == 0 {
		return fmt.Errorf("%s", "missing name")
	}

	return nil
}

type UserService interface {
	CreateUser(ctx context.Context, user *User) error
}
