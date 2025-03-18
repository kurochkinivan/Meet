package usecase

import (
	"context"

	"github.com/kurochkinivan/Meet/internal/entity"
	"github.com/sirupsen/logrus"
)

type UserUseCase struct {
	UserStorage
	UserCache
}

func NewUserUseCase(userStorage UserStorage, userCache UserCache) *UserUseCase {
	return &UserUseCase{
		UserStorage: userStorage,
		UserCache:   userCache,
	}
}

type UserStorage interface {
	GetByID(ctx context.Context, userID string) (*entity.User, error)
}

type UserCache interface {
	Get(ctx context.Context, userID string) (*entity.User, bool)
	Set(ctx context.Context, user *entity.User) error
}

func (u *UserUseCase) GetUserByID(ctx context.Context, userID string) (*entity.User, error) {
	if user, ok := u.UserCache.Get(ctx, userID); ok {
		return user, nil
	}

	user, err := u.UserStorage.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err = u.UserCache.Set(ctx, user); err != nil {
		logrus.WithError(err).Errorf("failed to set user for user %q", userID)
	}
	
	return user, nil
}
