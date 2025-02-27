package usecase

import "github.com/kurochkinivan/Meet/internal/usecase/repository/postgresql"

type UseCases struct {
	*AuthUseCase
	*PhotoUseCase
}

func NewUseCases(repositories *postgresql.Repositories) *UseCases {
	return &UseCases{
		AuthUseCase: NewAuthUseCase(repositories.UserRepository),
		PhotoUseCase: NewPhotoUseCase(repositories.PhotoRepository),
	}
}
