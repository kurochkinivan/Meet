package usecase

import (
	"github.com/kurochkinivan/Meet/config"
	"github.com/kurochkinivan/Meet/internal/usecase/repository/pg"
	"github.com/kurochkinivan/Meet/internal/usecase/repository/redis"
	"github.com/kurochkinivan/Meet/internal/usecase/repository/s3"
)

type UseCases struct {
	*PhotoUseCase
	*UserUseCase
}

func NewUseCases(cfg *config.Config, PGrepositories *pg.Repositories, S3Repositoires *s3.Repositories, redisRepositories *redis.Repositories) *UseCases {
	return &UseCases{
		PhotoUseCase: NewPhotoUseCase(PGrepositories.PhotoRepository, S3Repositoires.PhotoRepository, redisRepositories.UserRepository, int(cfg.S3.PhotoLimit)),
		UserUseCase:  NewUserUseCase(PGrepositories.UserRepository, redisRepositories.UserRepository),
	}
}
