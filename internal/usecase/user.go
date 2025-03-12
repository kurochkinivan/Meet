package usecase

import (
	"context"

	"github.com/kurochkinivan/Meet/internal/entity"
)

type UserUseCase struct {
	UserRepository
}

func NewUserUseCase(userRepository UserRepository) *UserUseCase {
	return &UserUseCase{
		UserRepository: userRepository,
	}
}

type UserRepository interface {
	GetByID(ctx context.Context, userID string) (*entity.User, error)
}

func (u *UserUseCase) GetUserByID(ctx context.Context, userID string) (*entity.User, error) {
	user, err := u.UserRepository.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}
