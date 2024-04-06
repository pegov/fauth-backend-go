package service

import "github.com/pegov/fauth-backend-go/internal/model"

type AuthService interface {
	Login(request *model.LoginRequest) (*model.User, error)
}

type authService struct{}

func NewAuthService() *authService {
	return &authService{}
}

func (s *authService) Login(request *model.LoginRequest) (*model.User, error) {
	return &model.User{
		ID:       1,
		Username: request.Login,
	}, nil
}
