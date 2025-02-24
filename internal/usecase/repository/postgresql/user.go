package postgresql

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
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
		return psql.ErrCreateQuery(op, err)
	}

	commTag, err := r.client.Exec(ctx, sql, args...)
	if err != nil {
		return psql.ErrExec(op, err)
	}

	if commTag.RowsAffected() == 0 {
		return psql.NoRowsAffected
	}

	return nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	op := "GetUserByEmail"

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
		return nil, psql.ErrCreateQuery(op, err)
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
		return nil, psql.ErrDoQuery(op, err)
	}

	return user, nil
}
