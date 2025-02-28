package pg

import (
	"context"
	"net/http"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kurochkinivan/Meet/internal/apperr"
	"github.com/kurochkinivan/Meet/pkg/psql"
)

type PhotoRepository struct {
	client *pgxpool.Pool
	qb         sq.StatementBuilderType
}

func NewPhotoRepository(client *pgxpool.Pool) *PhotoRepository {
	return &PhotoRepository{
		client: client,
		qb:         sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *PhotoRepository) CreatePhoto(ctx context.Context, userID string, url string) error {
	op := "CreatePhoto"

	sql, args, err := r.qb.
		Insert(TablePhotos).
		Columns(
			"user_id",
			"url",
		).
		Values(
			userID,
			url,
		).
		ToSql()
	if err != nil {
		return apperr.WithHTTPStatus(psql.ErrCreateQuery(op, err), http.StatusInternalServerError)
	}

	commTag, err := r.client.Exec(ctx, sql, args...)
	if err != nil {
		return apperr.WithHTTPStatus(psql.ErrExec(op, err), http.StatusInternalServerError)
	}

	if commTag.RowsAffected() == 0 {
		return apperr.WithHTTPStatus(psql.ErrNoRowsAffected, http.StatusInternalServerError)
	}

	return nil
}
