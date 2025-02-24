package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/kurochkinivan/Meet/internal/entity"
	"github.com/kurochkinivan/Meet/pkg/psql"
)

type AuthUseCase struct {
	UserRepository
}

func NewAuthUseCase(userRepository UserRepository) *AuthUseCase {
	return &AuthUseCase{
		UserRepository: userRepository,
	}
}

type UserRepository interface {
	CreateIfNotExists(ctx context.Context, user *entity.User) error
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
}

func (a *AuthUseCase) Register(ctx context.Context, user *entity.User) (*entity.User, error) {
	err := a.UserRepository.CreateIfNotExists(ctx, user)
	if err != nil {
		if errors.Is(err, psql.NoRowsAffected) {
			return nil, fmt.Errorf("user already exists")
		}
		return nil, fmt.Errorf("failed to create user, err: %w", err)
	}

	user, err = a.UserRepository.GetUserByEmail(ctx, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email, err: %w", err)
	}

	return user, nil
}
