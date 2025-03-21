package usecase

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/kurochkinivan/Meet/internal/apperr"
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
	Exists(ctx context.Context, phone string) (bool, error)
	CreateIfNotExists(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, userID string) (*entity.User, error)
	GetByPhone(ctx context.Context, phone string) (*entity.User, error)
	GetIfExists(ctx context.Context, phone, password string) (*entity.User, error)
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
	exists, err := u.UserStorage.Exists(ctx, user.Phone)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, apperr.ErrUserExists
	}

	err = u.UserStorage.CreateIfNotExists(ctx, user)
	if err != nil {
		return nil, err
	}

	user, err = u.UserStorage.GetByPhone(ctx, user.Phone)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *UserUseCase) AuthenticatePhone(ctx context.Context, phone, password string) (*entity.User, error) {
	user, err := u.UserStorage.GetIfExists(ctx, phone, u.hashPassword(password))
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *UserUseCase) AuthenticateOAuth(ctx context.Context, OAuth string) (*entity.User, error) {
	yandexResponse, err := yandexoauth.ParseOAuthToken(ctx, OAuth)
	if err != nil {
		return nil, err
	}

	birthday, err := time.Parse(time.DateOnly, yandexResponse.Birthday)
	if err != nil {
		return nil, fmt.Errorf("failed to parse birthday: %w", err)
	}

	user := &entity.User{
		Name:     yandexResponse.FirstName,
		BirthDay: birthday,
		Sex:      yandexResponse.Sex,
		Phone:    yandexResponse.Phone.Number,
	}

	err = u.UserStorage.CreateIfNotExists(ctx, user)
	if err != nil {
		return nil, err
	}

	user, err = u.UserStorage.GetByPhone(ctx, user.Phone)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *UserUseCase) hashPassword(password string) string {
	h := sha256.New()
	h.Write([]byte(password))
	return fmt.Sprintf("%x", h.Sum(nil))
}
