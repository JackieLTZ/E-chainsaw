package repos

import (
	"context"
	"database/sql"
	"errors"
)

type UserRepository struct {
	db *sql.DB
}

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func NewUserRepository(db *sql.DB) UsersInterface {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetUsers(ctx context.Context) ([]User, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name, email FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *UserRepository) GetUser(ctx context.Context, userID int) (*User, error) {
	var user User

	err := r.db.QueryRowContext(
		ctx,
		"SELECT * FROM users WHERE id = $1",
		userID,
	).Scan(&user.ID, &user.Name, &user.Email)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) DeleteUser(ctx context.Context, userID int) error {

	result, err := r.db.ExecContext(
		ctx,
		"DELETE FROM users WHERE id = $1",
		userID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

func (r *UserRepository) CreateUser(ctx context.Context, name, email string) (*User, error) {
	var user User
	err := r.db.QueryRowContext(
		ctx,
		"INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id, name, email",
		name,
		email,
	).Scan(&user.ID, &user.Name, &user.Email)

	if err != nil {
		return nil, err
	}

	return &user, nil
}
