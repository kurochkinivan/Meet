package pg

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repositories struct {
	UserRepository  *UserRepository
	PhotoRepository *PhotoRepository
}

func NewRepositories(client *pgxpool.Pool) *Repositories {
	return &Repositories{
		UserRepository:  NewUserRepository(client),
		PhotoRepository: NewPhotoRepository(client),
	}
}
