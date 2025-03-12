package pg

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kurochkinivan/Meet/internal/apperr"
	"github.com/kurochkinivan/Meet/internal/entity"
	"github.com/kurochkinivan/Meet/pkg/psql"
)

type UserRepository struct {
	client *pgxpool.Pool
	qb     sq.StatementBuilderType
}

func NewUserRepository(client *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		client: client,
		qb:     sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *UserRepository) CreateIfNotExists(ctx context.Context, user *entity.User) error {
	op := "CreateIfNotExists"

	sql, args, err := r.qb.
		Insert(TableUsers).
		Columns(
			"name",
			"email",
			"password",
			"location",
		).
		Values(
			user.Name,
			user.Email,
			user.Password,
			sq.Expr("ST_SetSRID(ST_MakePoint(?, ?), 4326)", user.Location.Longitude, user.Location.Latitude),
		).
		Suffix("ON CONFLICT (email) DO NOTHING").
		ToSql()
	if err != nil {
		return apperr.WithHTTPStatus(psql.ErrCreateQuery(op, err), http.StatusInternalServerError)
	}

	commTag, err := r.client.Exec(ctx, sql, args...)
	if err != nil {
		return apperr.WithHTTPStatus(psql.ErrExec(op, err), http.StatusInternalServerError)
	}

	if commTag.RowsAffected() == 0 {
		return apperr.WithHTTPStatus(psql.ErrNoRowsAffected, http.StatusConflict)
	}

	return nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	op := "GetByEmail"

	sql, args, err := r.qb.
		Select(
			"id",
			"name",
			"email",
			"ST_X(location::geometry) AS longitude",
			"ST_Y(location::geometry) AS latitude",
			"created_at",
		).
		From(TableUsers).
		Where(sq.Eq{"email": email}).
		ToSql()
	if err != nil {
		return nil, apperr.WithHTTPStatus(psql.ErrCreateQuery(op, err), http.StatusInternalServerError)
	}

	user := &entity.User{}
	err = r.client.QueryRow(ctx, sql, args...).Scan(
		&user.UUID,
		&user.Name,
		&user.Email,
		&user.Location.Longitude,
		&user.Location.Latitude,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, apperr.WithHTTPStatus(psql.ErrScan(op, err), http.StatusInternalServerError)
	}

	return user, nil
}

func (r *UserRepository) GetUserIfExists(ctx context.Context, email, password string) (*entity.User, error) {
	op := "GetUserIfExists"

	sql, args, err := r.qb.
		Select(
			"id",
			"name",
			"email",
			"ST_X(location::geometry) AS longitude",
			"ST_Y(location::geometry) AS latitude",
			"created_at",
		).
		From(TableUsers).
		Where(sq.And{sq.Eq{"email": email}, sq.Eq{"password": password}}).
		ToSql()
	if err != nil {
		return nil, apperr.WithHTTPStatus(psql.ErrCreateQuery(op, err), http.StatusInternalServerError)
	}

	user := &entity.User{}
	err = r.client.QueryRow(ctx, sql, args...).Scan(
		&user.UUID,
		&user.Name,
		&user.Email,
		&user.Location.Longitude,
		&user.Location.Latitude,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.WithHTTPStatus(err, http.StatusUnauthorized)
		}
		return nil, psql.ErrScan(op, err)
	}

	return user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, userID string) (*entity.User, error) {
	op := "GetByID"

	sql, args, err := r.qb.
		Select(
			usersField("id"),
			usersField("name"),
			usersField("email"),
			"ST_X(users.location::geometry) AS longitude",
			"ST_Y(users.location::geometry) AS latitude",
			usersField("created_at"),
			photosField("id"),
			photosField("url"),
		).
		From(TableUsers).
		LeftJoin(fmt.Sprintf("%s ON %s.id = %s.user_id", TablePhotos, TableUsers, TablePhotos)).
		Where(sq.Eq{usersField("id"): userID}).
		ToSql()
	if err != nil {
		return nil, apperr.WithHTTPStatus(psql.ErrCreateQuery(op, err), http.StatusInternalServerError)
	}

	user := &entity.User{}
	rows, err := r.client.Query(ctx, sql, args...)
	if err != nil {
		return nil, apperr.WithHTTPStatus(psql.ErrDoQuery(op, err), http.StatusInternalServerError)
	}

	for rows.Next() {
		photo := &entity.Photo{}
		err = rows.Scan(
			&user.UUID,
			&user.Name,
			&user.Email,
			&user.Location.Longitude,
			&user.Location.Latitude,
			&user.CreatedAt,
			&photo.ID,
			&photo.URL,
		)
		if err != nil {
			return nil, apperr.WithHTTPStatus(psql.ErrScan(op, err), http.StatusInternalServerError)
		}

		user.Photos = append(user.Photos, photo)
	}

	return user, nil
}
