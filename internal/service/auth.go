package service

import (
	"errors"
	"strings"

	"github.com/pegov/fauth-backend-go/internal/captcha"
	"github.com/pegov/fauth-backend-go/internal/model"
	"github.com/pegov/fauth-backend-go/internal/password"
	"github.com/pegov/fauth-backend-go/internal/repo"
	"github.com/pegov/fauth-backend-go/internal/token"
)

type AuthService interface {
	Register(request *model.RegisterRequest) (string, string, error)
	Login(request *model.LoginRequest) (string, string, error)
	Token(accessToken string) (*token.User, error)
	RefreshToken(refreshToken string) (string, error)
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
	ErrUserNotActive             = errors.New("user not active") // 401
	ErrUserPasswordNotSet        = errors.New("password not set")
	ErrPasswordVerification      = errors.New("user password verification") // 401
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

func (s *authService) Login(request *model.LoginRequest) (string, string, error) {
	login := strings.TrimSpace(request.Login)
	user, err := s.userRepo.GetByLogin(login)
	if err != nil {
		return "", "", ErrUserNotFound
	}

	if !user.Active {
		return "", "", ErrUserNotActive
	}

	if user.Password == nil {
		return "", "", ErrUserPasswordNotSet
	}

	if s.passwordHasher.Compare([]byte(*user.Password), []byte(request.Password)) != nil {
		return "", "", ErrPasswordVerification
	}

	s.userRepo.UpdateLastLogin(user.ID)

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

var (
	ErrTokenDecoding = errors.New("token decoding error")
)

func (s *authService) Token(accessToken string) (*token.User, error) {
	user, err := s.tokenBackend.Decode(accessToken, token.AccessTokenType)
	if err != nil {
		return nil, ErrTokenDecoding
	}

	return token.UserPayloadFromUserClaims(user), nil
}

func (s *authService) RefreshToken(refreshToken string) (string, error) {
	refreshTokenClaims, err := s.tokenBackend.Decode(refreshToken, token.AccessTokenType)
	if err != nil {
		return "", ErrTokenDecoding
	}

	// TODO: check ban and kick list, mass ban ts

	user, err := s.userRepo.Get(refreshTokenClaims.ID)
	if err != nil {
		return "", err
	}

	payload := token.User{
		ID:       user.ID,
		Username: user.Username,
		Roles:    []string{},
	}
	a, err := s.tokenBackend.Encode(&payload, 60*60*6, token.AccessTokenType)
	if err != nil {
		return "", err
	}

	return a, nil
}
