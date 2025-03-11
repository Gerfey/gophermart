package mock_service

import (
	"context"
)

type UserServiceMock interface {
	RegisterUser(ctx context.Context, login, password string) (string, error)
	LoginUser(ctx context.Context, login, password string) (string, error)
	ParseToken(token string) (int64, error)
}
