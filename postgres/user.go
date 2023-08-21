package postgres

import (
	"context"
	"fmt"
	"strings"

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

func (s *UserService) FindUserByID(ctx context.Context, id int) (user *app.User, err error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Fetch user
	user, err = findUserByID(ctx, tx, id)
	if err != nil {
		return nil, err
	}

	return user, nil
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

// findUserByID is a helper function to fetch a user by ID.
// Returns ENOTFOUND if user does not exist.
func findUserByID(ctx context.Context, tx *Tx, id int) (*app.User, error) {
	a, _, err := findUsers(ctx, tx, app.UserFilter{ID: &id})
	if err != nil {
		return nil, err
	} else if len(a) == 0 {
		return nil, fmt.Errorf("User not found. Err: %v", err)
	}
	return a[0], nil
}

// findUsers returns a list of users matching a filter. Also returns a count of
// total matching users which may differ if filter.Limit is set.
func findUsers(ctx context.Context, tx *Tx, filter app.UserFilter) (_ []*app.User, n int, err error) {
	// Build WHERE clause.
	where, args := []string{"1 = 1"}, []interface{}{}
	if v := filter.ID; v != nil {
		where, args = append(where, "id = ?"), append(args, *v)
	}

	// Execute query to fetch user rows.
	rows, err := tx.QueryContext(ctx, `
		SELECT 
		    id,
		    name,
		    created_at,
		    updated_at,
		    COUNT(*) OVER()
		FROM users
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY id ASC
		`+FormatLimitOffset(filter.Limit, filter.Offset),
		args...,
	)
	if err != nil {
		return nil, n, err
	}
	defer rows.Close()

	// Deserialize rows into User objects.
	users := make([]*app.User, 0)
	for rows.Next() {
		var user app.User
		if err := rows.Scan(
			&user.ID,
			&user.Name,
			(*NullTime)(&user.CreatedAt),
			(*NullTime)(&user.UpdatedAt),
			&n,
		); err != nil {
			return nil, 0, err
		}

		users = append(users, &user)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return users, n, nil
}
