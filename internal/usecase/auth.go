package usecase

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/kurochkinivan/Meet/internal/entity"
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
	GetUserIfExists(ctx context.Context, email, password string) (*entity.User, error)
}

func (u *AuthUseCase) Register(ctx context.Context, user *entity.User) (*entity.User, error) {
	user.Password = u.hashPassword(user.Password)
	err := u.UserRepository.CreateIfNotExists(ctx, user)
	if err != nil {
		return nil, err
	}

	user, err = u.UserRepository.GetUserByEmail(ctx, user.Email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *AuthUseCase) AuthenticateUser(ctx context.Context, email, password string) (*entity.User, error) {
	user, err := u.UserRepository.GetUserIfExists(ctx, email, u.hashPassword(password))
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *AuthUseCase) hashPassword(password string) string {
	h := sha256.New()
	h.Write([]byte(password))
	return fmt.Sprintf("%x", h.Sum(nil))
}
