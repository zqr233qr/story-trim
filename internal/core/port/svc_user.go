package port

import (
	"context"
)

type UserService interface {
	Register(ctx context.Context, username, password string) error
	Login(ctx context.Context, username, password string) (string, error)
	ValidateToken(token string) (uint, error)
}
