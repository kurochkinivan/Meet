package postgresql

import "github.com/jackc/pgx/v5/pgxpool"

type Repositories struct {
	UserRepository *UserRepository
}

func NewRepositories(client *pgxpool.Pool) *Repositories {
	return &Repositories{
		UserRepository: NewUserRepository(client),
	}
}
