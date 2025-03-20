package usecase

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/kurochkinivan/Meet/internal/entity"
	yandexoauth "github.com/kurochkinivan/Meet/internal/external/yandexOAuth"
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
	CreateIfNotExists(ctx context.Context, user *entity.User) error
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetIfExists(ctx context.Context, email, password string) (*entity.User, error)
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

func (u *UserUseCase) Register(ctx context.Context, user *entity.User) (*entity.User, error) {
	user.Password = u.hashPassword(user.Password)
	err := u.UserStorage.CreateIfNotExists(ctx, user)
	if err != nil {
		return nil, err
	}

	user, err = u.UserStorage.GetByEmail(ctx, user.Email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *UserUseCase) AuthenticateEmail(ctx context.Context, email, password string) (*entity.User, error) {
	user, err := u.UserStorage.GetIfExists(ctx, email, u.hashPassword(password))
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *UserUseCase) AuthenticateOAuth(ctx context.Context, OAuth string) (*entity.User, error) {
	err := yandexoauth.GetInfoByToken(ctx, OAuth)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (u *UserUseCase) hashPassword(password string) string {
	h := sha256.New()
	h.Write([]byte(password))
	return fmt.Sprintf("%x", h.Sum(nil))
}
