package usecase

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/kurochkinivan/Meet/internal/entity"
)

type AuthUseCase struct {
	UserCreator
}

func NewAuthUseCase(userCreatorRepository UserCreator) *AuthUseCase {
	return &AuthUseCase{
		UserCreator: userCreatorRepository,
	}
}

type UserCreator interface {
	CreateIfNotExists(ctx context.Context, user *entity.User) error
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetUserIfExists(ctx context.Context, email, password string) (*entity.User, error)
}

func (u *AuthUseCase) Register(ctx context.Context, user *entity.User) (*entity.User, error) {
	user.Password = u.hashPassword(user.Password)
	err := u.UserCreator.CreateIfNotExists(ctx, user)
	if err != nil {
		return nil, err
	}

	user, err = u.UserCreator.GetByEmail(ctx, user.Email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *AuthUseCase) AuthenticateUser(ctx context.Context, email, password string) (*entity.User, error) {
	user, err := u.UserCreator.GetUserIfExists(ctx, email, u.hashPassword(password))
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
