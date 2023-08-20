package postgres

import (
	"context"
	"fmt"

	app "github.com/truescotian/golang-rest-projstructure"
)

// Ensure service implemented interface
var _ app.UserService = (*UserService)(nil)

type UserService struct {
	db *DB
}

func NewUserService(db *DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) CreateUser(ctx context.Context, user *app.User) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := createUser(ctx, tx, user); err != nil {
		return err
	}

	return tx.Commit()
}

func createUser(ctx context.Context, tx *Tx, user *app.User) error {
	if err := user.Validate(); err != nil {
		return err
	}

	query := fmt.Sprintf(`INSERT INTO users (username) VALUES ('%s')`, user.Name)

	result, err := tx.ExecContext(
		ctx,
		query,
	)

	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	user.ID = int(id)

	return nil
}
