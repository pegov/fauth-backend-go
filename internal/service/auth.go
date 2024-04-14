package service

import (
	"errors"

	"github.com/pegov/fauth-backend-go/internal/captcha"
	"github.com/pegov/fauth-backend-go/internal/model"
	"github.com/pegov/fauth-backend-go/internal/password"
	"github.com/pegov/fauth-backend-go/internal/repo"
	"github.com/pegov/fauth-backend-go/internal/token"
)

type AuthService interface {
	Register(request *model.RegisterRequest) (string, string, error)
}

type authService struct {
	userRepo       repo.UserRepo
	captchaClient  captcha.CaptchaClient
	passwordHasher password.PasswordHasher
	tokenBackend   token.JwtBackend
}

func NewAuthService(
	userRepo repo.UserRepo,
	captchaClient captcha.CaptchaClient,
	passwordHasher password.PasswordHasher,
	tokenBackend token.JwtBackend,
) *authService {
	return &authService{
		userRepo:       userRepo,
		captchaClient:  captchaClient,
		passwordHasher: passwordHasher,
		tokenBackend:   tokenBackend,
	}
}

var (
	ErrUserNotFound              = errors.New("user not found")
	ErrUserAlreadyExistsEmail    = errors.New("email already exists")
	ErrUserAlreadyExistsUsername = errors.New("username already exists")
	ErrInvalidCaptcha            = errors.New("invalid captcha")
)

func (s *authService) Register(request *model.RegisterRequest) (string, string, error) {
	if err := request.Validate(); err != nil {
		return "", "", err
	}

	if !s.captchaClient.IsValid(request.Captcha) {
		return "", "", ErrInvalidCaptcha
	}

	user, err := s.userRepo.GetByEmail(request.Email)
	if err != nil {
		return "", "", err
	}

	if user != nil {
		return "", "", ErrUserAlreadyExistsEmail
	}

	user, err = s.userRepo.GetByUsername(request.Username)
	if err != nil {
		return "", "", err
	}

	if user != nil {
		return "", "", ErrUserAlreadyExistsUsername
	}

	passwordHash, err := s.passwordHasher.Hash([]byte(request.Password1))
	if err != nil {
		panic(err) // if password > 72 bytes (per bcrypt docs)
	}

	userCreate := model.UserCreate{
		Email:    request.Email,
		Username: request.Username,
		Password: string(passwordHash),
		Verified: false,
	}

	id, err := s.userRepo.Create(&userCreate)
	if err != nil {
		return "", "", err
	}

	// user != nil
	user, err = s.userRepo.Get(id)
	if err != nil {
		return "", "", err
	}

	payload := token.User{
		ID:       user.ID,
		Username: user.Username,
		Roles:    []string{},
	}
	a, err := s.tokenBackend.Encode(&payload, 60*60*6, token.AccessTokenType)
	if err != nil {
		return "", "", err
	}
	r, err := s.tokenBackend.Encode(&payload, 60*60*24*31, token.RefreshTokenType)
	if err != nil {
		return "", "", err
	}

	return a, r, nil
}
