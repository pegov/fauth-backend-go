package service

import (
	"errors"
	"fmt"

	"github.com/pegov/fauth-backend-go/internal/model"
	"github.com/pegov/fauth-backend-go/internal/repo"
)

type AuthService interface {
	Login(request *model.LoginRequest) (*model.User, error)
	Get(id int) (*model.User, error)
}

type authService struct {
	userRepo repo.UserRepo
}

func NewAuthService(userRepo repo.UserRepo) *authService {
	return &authService{
		userRepo: userRepo,
	}
}

var (
	ErrUserNotFound = errors.New("user not found")
)

func (s *authService) Login(request *model.LoginRequest) (*model.User, error) {
	return &model.User{
		ID:       1,
		Username: request.Login,
	}, nil
}

func (s *authService) Get(id int) (*model.User, error) {
	user, err := s.userRepo.Get(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id from db: %w", err)
	}

	if user == nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}
