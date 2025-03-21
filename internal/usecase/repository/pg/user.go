package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kurochkinivan/Meet/internal/apperr"
	"github.com/kurochkinivan/Meet/internal/entity"
	pgclient "github.com/kurochkinivan/Meet/pkg/pgClient"
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
			"birthday",
			"sex",
			"phone",
			"password",
			"location",
		).
		Values(
			user.Name,
			user.BirthDay,
			user.Sex,
			user.Phone,
			user.Password,
			sq.Expr("ST_SetSRID(ST_MakePoint(?, ?), 4326)", user.Location.Longitude, user.Location.Latitude),
		).
		Suffix("ON CONFLICT (phone) DO NOTHING").
		ToSql()
	if err != nil {
		return apperr.WithHTTPStatus(pgclient.ErrCreateQuery(op, err), http.StatusInternalServerError)
	}

	_, err = r.client.Exec(ctx, sql, args...)
	if err != nil {
		return apperr.WithHTTPStatus(pgclient.ErrExec(op, err), http.StatusInternalServerError)
	}

	return nil
}

func (r *UserRepository) GetByPhone(ctx context.Context, phone string) (*entity.User, error) {
	op := "GetByphone"

	sql, args, err := r.qb.
		Select(
			"id",
			"name",
			"birthday",
			"sex",
			"phone",
			"ST_X(location::geometry) AS longitude",
			"ST_Y(location::geometry) AS latitude",
			"created_at",
		).
		From(TableUsers).
		Where(sq.Eq{"phone": phone}).
		ToSql()
	if err != nil {
		return nil, apperr.WithHTTPStatus(pgclient.ErrCreateQuery(op, err), http.StatusInternalServerError)
	}

	var user entity.User
	err = r.client.QueryRow(ctx, sql, args...).Scan(
		&user.UUID,
		&user.Name,
		&user.BirthDay,
		&user.Sex,
		&user.Phone,
		&user.Location.Longitude,
		&user.Location.Latitude,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, apperr.WithHTTPStatus(pgclient.ErrScan(op, err), http.StatusInternalServerError)
	}

	return &user, nil
}

func (r *UserRepository) GetIfExists(ctx context.Context, phone, password string) (*entity.User, error) {
	op := "GetUserIfExists"

	sql, args, err := r.qb.
		Select(
			"id",
			"name",
			"birthday",
			"sex",
			"phone",
			"ST_X(location::geometry) AS longitude",
			"ST_Y(location::geometry) AS latitude",
			"created_at",
		).
		From(TableUsers).
		Where(sq.And{sq.Eq{"phone": phone}, sq.Eq{"password": password}}).
		ToSql()
	if err != nil {
		return nil, apperr.WithHTTPStatus(pgclient.ErrCreateQuery(op, err), http.StatusInternalServerError)
	}

	var user entity.User
	err = r.client.QueryRow(ctx, sql, args...).Scan(
		&user.UUID,
		&user.Name,
		&user.BirthDay,
		&user.Sex,
		&user.Phone,
		&user.Location.Longitude,
		&user.Location.Latitude,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.WithHTTPStatus(apperr.ErrNoRows, http.StatusUnauthorized)
		}
		return nil, pgclient.ErrScan(op, err)
	}

	return &user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, userID string) (*entity.User, error) {
	op := "GetByID"

	query, args, err := r.qb.
		Select(
			usersField("id"),
			usersField("name"),
			usersField("birthday"),
			usersField("sex"),
			usersField("phone"),
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
		return nil, apperr.WithHTTPStatus(pgclient.ErrCreateQuery(op, err), http.StatusInternalServerError)
	}

	rows, err := r.client.Query(ctx, query, args...)
	if err != nil {
		return nil, apperr.WithHTTPStatus(pgclient.ErrDoQuery(op, err), http.StatusInternalServerError)
	}

	user := &entity.User{Photos: make([]*entity.Photo, 0)}
	for rows.Next() {
		var photoID sql.NullInt64
		var photoURL sql.NullString

		err = rows.Scan(
			&user.UUID,
			&user.Name,
			&user.BirthDay,
			&user.Sex,
			&user.Phone,
			&user.Location.Longitude,
			&user.Location.Latitude,
			&user.CreatedAt,
			&photoID,
			&photoURL,
		)
		if err != nil {
			return nil, apperr.WithHTTPStatus(pgclient.ErrScan(op, err), http.StatusInternalServerError)
		}

		if photoID.Valid {
			user.Photos = append(user.Photos, &entity.Photo{
				ID:  photoID.Int64,
				URL: photoURL.String,
			})
		}

	}

	return user, nil
}

func (r *UserRepository) Exists(ctx context.Context, phone string) (bool, error) {
	op := "Exists"

	sql, args, err := r.qb.
		Select("1").
		Prefix("SELECT EXISTS (").
		From(TableUsers).
		Where(sq.Eq{"phone": phone}).
		Suffix(")").
		ToSql()
	if err != nil {
		return false, apperr.WithHTTPStatus(pgclient.ErrCreateQuery(op, err), http.StatusInternalServerError)
	}

	var exists bool
	err = r.client.QueryRow(ctx, sql, args...).Scan(&exists)
	if err != nil {
		return false, apperr.WithHTTPStatus(pgclient.ErrScan(op, err), http.StatusInternalServerError)
	}

	return exists, nil
}
