package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

	"github.com/pegov/fauth-backend-go/internal/entity"
	"github.com/pegov/fauth-backend-go/internal/model"
)

type UserRepo interface {
	Get(ctx context.Context, id int32) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetByUsername(ctx context.Context, username string) (*entity.User, error)
	GetByLogin(ctx context.Context, login string) (*entity.User, error)
	Create(ctx context.Context, data *model.UserCreate) (int32, error)
	UpdateLastLogin(ctx context.Context, id int32) error
	Ban(ctx context.Context, id int32) error
	Unban(ctx context.Context, id int32) error
	Kick(ctx context.Context, id int32) error
	Unkick(ctx context.Context, id int32) error
	WasRecentlyBanned(ctx context.Context, id int32) (bool, error)
	IsKicked(ctx context.Context, id int32) (bool, error)
}

type userRepo struct {
	db    *sqlx.DB
	cache *redis.Client
}

func NewUserRepo(db *sqlx.DB, cache *redis.Client) UserRepo {
	return &userRepo{db: db, cache: cache}
}

func (r *userRepo) Create(ctx context.Context, data *model.UserCreate) (int32, error) {
	var id int32
	now := time.Now().UTC()
	if err := r.db.Get(&id, `
		INSERT INTO auth_user(
			email, username, password, active, verified, created_at, last_login
		) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id
	`, data.Email, data.Username, data.Password, true, data.Verified, now, now); err != nil {
		return 0, err
	}

	return id, nil
}

func (r *userRepo) Get(ctx context.Context, id int32) (*entity.User, error) {
	var user entity.User
	if err := r.db.Get(&user, `
		SELECT id, email, username, password, active, verified, created_at, last_login FROM auth_user WHERE id = $1
	`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &user, nil
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	if err := r.db.Get(&user, `
		SELECT id, email, username, password, active, verified, created_at, last_login FROM auth_user WHERE email = $1
	`, email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &user, nil
}

func (r *userRepo) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	var user entity.User
	if err := r.db.Get(&user, `
		SELECT id, email, username, password, active, verified, created_at, last_login FROM auth_user WHERE username = $1
	`, username); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &user, nil
}

func (r *userRepo) GetByLogin(ctx context.Context, login string) (*entity.User, error) {
	if strings.Contains(login, "@") {
		user, err := r.GetByEmail(ctx, login)
		if err != nil {
			return nil, err
		}

		if user != nil {
			return user, nil
		}
	}

	return r.GetByUsername(ctx, login)
}

func (r *userRepo) UpdateLastLogin(ctx context.Context, id int32) error {
	now := time.Now().UTC()
	_, err := r.db.Exec("UPDATE auth_user SET last_login = $1", now)
	return err
}

func (r *userRepo) Ban(ctx context.Context, id int32) error {
	_, err := r.db.ExecContext(ctx, "UPDATE auth_user SET active = false WHERE id = $1", id)
	return err
}

func (r *userRepo) Unban(ctx context.Context, id int32) error {
	_, err := r.db.ExecContext(ctx, "UPDATE auth_user SET active = true WHERE id = $1", id)
	return err
}

func (r *userRepo) Kick(ctx context.Context, id int32) error {
	ts := time.Now().UTC().Unix()
	key := fmt.Sprintf("users:kick:%d", id)
	return r.cache.Set(ctx, key, ts, 60*60*6*time.Second).Err()
}

func (r *userRepo) Unkick(ctx context.Context, id int32) error {
	key := fmt.Sprintf("users:kick:%d", id)
	return r.cache.Del(ctx, key).Err()
}

func (r *userRepo) WasRecentlyBanned(ctx context.Context, id int32) (bool, error) {
	key := fmt.Sprintf("users:ban:%d", id)
	if err := r.cache.Get(ctx, key).Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (r *userRepo) IsKicked(ctx context.Context, id int32) (bool, error) {
	key := fmt.Sprintf("users:kick:%d", id)
	if err := r.cache.Get(ctx, key).Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
