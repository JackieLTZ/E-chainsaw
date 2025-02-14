package repos

import "context"

type UsersInterface interface {
	GetUser(ctx context.Context, userID int) (*User, error)
	GetUsers(ctx context.Context) ([]User, error)
	CreateUser(ctx context.Context, name string, email string) (*User, error)
	DeleteUser(ctx context.Context, userID int) error
}
