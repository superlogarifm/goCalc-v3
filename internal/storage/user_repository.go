package storage

import (
	"context"

	"github.com/superlogarifm/goCalc-v3/internal/models"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)
}
