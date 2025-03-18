package usecase

import (
	"github.com/kurochkinivan/Meet/internal/usecase/repository/pg"
	"github.com/kurochkinivan/Meet/internal/usecase/repository/redis"
	"github.com/kurochkinivan/Meet/internal/usecase/repository/s3"
)

type UseCases struct {
	*AuthUseCase
	*PhotoUseCase
	*UserUseCase
}

func NewUseCases(PGrepositories *pg.Repositories, S3Repositoires *s3.Repositories, redisRepositories *redis.Repositories) *UseCases {
	return &UseCases{
		AuthUseCase:  NewAuthUseCase(PGrepositories.UserRepository),
		PhotoUseCase: NewPhotoUseCase(PGrepositories.PhotoRepository, S3Repositoires.PhotoRepository),
		UserUseCase:  NewUserUseCase(PGrepositories.UserRepository, redisRepositories.UserRepository),
	}
}
