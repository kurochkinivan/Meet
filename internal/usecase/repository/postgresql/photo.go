package postgresql

import (
	"context"
	"mime/multipart"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PhotoRepository struct {
	client *pgxpool.Pool
	qb     sq.StatementBuilderType
}

func NewPhotoRepository(client *pgxpool.Pool) *PhotoRepository {
	return &PhotoRepository{
		client: client,
		qb:     sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *PhotoRepository) CreatePhotos(ctx context.Context, files []*multipart.FileHeader) {
	
}
