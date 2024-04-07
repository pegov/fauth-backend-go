package repo

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/pegov/fauth-backend-go/internal/model"
)

type UserRepo interface {
	Get(id int) (*model.User, error)
}

type userRepo struct {
	db *sqlx.DB
}

func NewUserRepo(db *sqlx.DB) UserRepo {
	return &userRepo{db: db}
}

func (r *userRepo) Get(id int) (*model.User, error) {
	var user model.User
	if err := r.db.Get(&user, "SELECT id, username FROM auth_user WHERE id = $1", id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &user, nil
}
