package repo

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/pegov/fauth-backend-go/internal/entity"
	"github.com/pegov/fauth-backend-go/internal/model"
)

type UserRepo interface {
	Get(id int32) (*entity.User, error)
	GetByEmail(email string) (*entity.User, error)
	GetByUsername(username string) (*entity.User, error)
	GetByLogin(login string) (*entity.User, error)
	Create(data *model.UserCreate) (int32, error)
}

type userRepo struct {
	db *sqlx.DB
}

func NewUserRepo(db *sqlx.DB) UserRepo {
	return &userRepo{db: db}
}

func (r *userRepo) Create(data *model.UserCreate) (int32, error) {
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

func (r *userRepo) Get(id int32) (*entity.User, error) {
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

func (r *userRepo) GetByEmail(email string) (*entity.User, error) {
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

func (r *userRepo) GetByUsername(username string) (*entity.User, error) {
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

func (r *userRepo) GetByLogin(login string) (*entity.User, error) {
	if strings.Contains(login, "@") {
		user, err := r.GetByEmail(login)
		if err != nil {
			return nil, err
		}

		if user != nil {
			return user, nil
		}
	}

	return r.GetByUsername(login)
}
