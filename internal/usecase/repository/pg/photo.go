package pg

import (
	"context"
	"errors"
	"net/http"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kurochkinivan/Meet/internal/apperr"
	"github.com/kurochkinivan/Meet/internal/entity"
	"github.com/kurochkinivan/Meet/pkg/pgClient"
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

func (r *PhotoRepository) CreatePhoto(ctx context.Context, userID string, url string, objectKey string) error {
	op := "CreatePhoto"

	sql, args, err := r.qb.
		Insert(TablePhotos).
		Columns(
			"user_id",
			"url",
			"object_key",
		).
		Values(
			userID,
			url,
			objectKey,
		).
		ToSql()
	if err != nil {
		return apperr.WithHTTPStatus(pgclient.ErrCreateQuery(op, err), http.StatusInternalServerError)
	}

	commTag, err := r.client.Exec(ctx, sql, args...)
	if err != nil {
		return apperr.WithHTTPStatus(pgclient.ErrExec(op, err), http.StatusInternalServerError)
	}

	if commTag.RowsAffected() == 0 {
		return apperr.WithHTTPStatus(pgclient.ErrNoRowsAffected, http.StatusInternalServerError)
	}

	return nil
}

func (r *PhotoRepository) GetPhotos(ctx context.Context, userID string) ([]*entity.Photo, error) {
	op := "GetPhotos"

	sql, args, err := r.qb.
		Select(
			"id",
			"user_id",
			"object_key",
			"url",
			"created_at",
		).
		From(TablePhotos).
		Where(sq.Eq{"user_id": userID}).
		ToSql()
	if err != nil {
		return nil, apperr.WithHTTPStatus(pgclient.ErrCreateQuery(op, err), http.StatusInternalServerError)
	}

	rows, err := r.client.Query(ctx, sql, args...)
	if err != nil {
		return nil, apperr.WithHTTPStatus(pgclient.ErrDoQuery(op, err), http.StatusInternalServerError)
	}

	photos := []*entity.Photo{}
	for rows.Next() {
		photo := &entity.Photo{}
		err = rows.Scan(
			&photo.ID,
			&photo.UserID,
			&photo.ObjectKey,
			&photo.URL,
			&photo.CreatedAt,
		)
		if err != nil {
			return nil, apperr.WithHTTPStatus(pgclient.ErrScan(op, err), http.StatusInternalServerError)
		}
		photos = append(photos, photo)
	}

	return photos, nil
}

func (r *PhotoRepository) DeletePhoto(ctx context.Context, userID string, photoID string) error {
	op := "DeletePhoto"

	sql, args, err := r.qb.
		Delete(TablePhotos).
		Where(sq.And{
			sq.Eq{"user_id": userID},
			sq.Eq{"id": photoID},
		}).
		ToSql()
	if err != nil {
		return apperr.WithHTTPStatus(pgclient.ErrCreateQuery(op, err), http.StatusInternalServerError)
	}

	commTag, err := r.client.Exec(ctx, sql, args...)
	if err != nil {
		return apperr.WithHTTPStatus(pgclient.ErrExec(op, err), http.StatusInternalServerError)
	}

	if commTag.RowsAffected() == 0 {
		return apperr.WithHTTPStatus(pgclient.ErrNoRowsAffected, http.StatusInternalServerError)
	}

	return nil
}

func (r *PhotoRepository) GetPhoto(ctx context.Context, photoID string) (*entity.Photo, error) {
	op := "GetPhoto"

	sql, args, err := r.qb.
		Select(
			"id",
			"user_id",
			"object_key",
			"url",
			"created_at",
		).
		From(TablePhotos).
		Where(sq.Eq{"id": photoID}).
		ToSql()
	if err != nil {
		return nil, apperr.WithHTTPStatus(pgclient.ErrCreateQuery(op, err), http.StatusInternalServerError)
	}

	photo := &entity.Photo{}
	err = r.client.QueryRow(ctx, sql, args...).Scan(
		&photo.ID,
		&photo.UserID,
		&photo.ObjectKey,
		&photo.URL,
		&photo.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.WithHTTPStatus(pgclient.ErrNoRows, http.StatusBadRequest)
		}
		return nil, apperr.WithHTTPStatus(pgclient.ErrScan(op, err), http.StatusInternalServerError)
	}

	return photo, nil
}
