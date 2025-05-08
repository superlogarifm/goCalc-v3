package postgres

import (
	"context"
	"errors"

	"github.com/superlogarifm/goCalc-v3/internal/models"
	"github.com/superlogarifm/goCalc-v3/internal/storage"

	"gorm.io/gorm"
)

type PGUserRepository struct {
	db *gorm.DB
}

func NewPGUserRepository(db *gorm.DB) *PGUserRepository {
	return &PGUserRepository{db: db}
}

func (r *PGUserRepository) CreateUser(ctx context.Context, user *models.User) error {
	result := r.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *PGUserRepository) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	var user models.User
	result := r.db.WithContext(ctx).Where("login = ?", login).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, storage.ErrUserNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

func (r *PGUserRepository) AutoMigrate() error {
	return r.db.AutoMigrate(&models.User{})
}
