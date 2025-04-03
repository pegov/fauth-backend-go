package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/pegov/fauth-backend-go/internal/captcha"
	"github.com/pegov/fauth-backend-go/internal/email"
	"github.com/pegov/fauth-backend-go/internal/model"
	"github.com/pegov/fauth-backend-go/internal/password"
	"github.com/pegov/fauth-backend-go/internal/repo"
	"github.com/pegov/fauth-backend-go/internal/token"
)

type AuthService interface {
	Register(ctx context.Context, request *model.RegisterRequest) (*Tokens, error)
	Login(ctx context.Context, request *model.LoginRequest) (*Tokens, error)
	Token(ctx context.Context, accessToken string) (*token.User, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, error)
	Me(ctx context.Context, id int32) (*model.Me, error)
}

type authService struct {
	userRepo       repo.UserRepo
	captchaClient  captcha.CaptchaClient
	passwordHasher password.PasswordManager
	tokenBackend   token.JwtBackend
	emailClient    email.EmailClient
}

func NewAuthService(
	userRepo repo.UserRepo,
	captchaClient captcha.CaptchaClient,
	passwordHasher password.PasswordManager,
	tokenBackend token.JwtBackend,
	emailClient email.EmailClient,
) *authService {
	return &authService{
		userRepo:       userRepo,
		captchaClient:  captchaClient,
		passwordHasher: passwordHasher,
		tokenBackend:   tokenBackend,
		emailClient:    emailClient,
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
	ErrUserWasKicked             = errors.New("user was kicked")
	ErrUserInMassLogout          = errors.New("user in mass logout")
)

type Tokens struct {
	Access  string
	Refresh string
}

func (s *authService) Register(
	ctx context.Context,
	request *model.RegisterRequest,
) (*Tokens, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}

	if !s.captchaClient.IsValid(request.Captcha) {
		return nil, ErrInvalidCaptcha
	}

	user, err := s.userRepo.GetByEmail(ctx, request.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	if user != nil {
		return nil, ErrUserAlreadyExistsEmail
	}

	user, err = s.userRepo.GetByUsername(ctx, request.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	if user != nil {
		return nil, ErrUserAlreadyExistsUsername
	}

	passwordHash, err := s.passwordHasher.Hash([]byte(request.Password1))
	if err != nil {
		return nil, fmt.Errorf("unexpected err: %w", err) // if password > 72 bytes (per bcrypt docs)
	}

	userCreate := model.UserCreate{
		Email:    request.Email,
		Username: request.Username,
		Password: string(passwordHash),
		Verified: false,
	}

	id, err := s.userRepo.Create(ctx, &userCreate)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// user != nil
	user, err = s.userRepo.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	payload := token.User{
		ID:       user.ID,
		Username: user.Username,
		Roles:    []string{},
	}
	a, err := s.tokenBackend.Encode(&payload, 60*60*6, token.AccessTokenType)
	if err != nil {
		return nil, err
	}
	r, err := s.tokenBackend.Encode(&payload, 60*60*24*31, token.RefreshTokenType)
	if err != nil {
		return nil, err
	}

	return &Tokens{a, r}, nil
}

func (s *authService) Login(
	ctx context.Context,
	request *model.LoginRequest,
) (*Tokens, error) {
	login := strings.TrimSpace(request.Login)
	user, err := s.userRepo.GetByLogin(ctx, login)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if !user.Active {
		return nil, ErrUserNotActive
	}

	if user.Password == nil {
		return nil, ErrUserPasswordNotSet
	}

	if s.passwordHasher.Compare(
		[]byte(*user.Password),
		[]byte(request.Password),
	) != nil {
		return nil, ErrPasswordVerification
	}

	s.userRepo.UpdateLastLogin(ctx, user.ID)

	payload := token.User{
		ID:       user.ID,
		Username: user.Username,
		Roles:    []string{},
	}
	a, err := s.tokenBackend.Encode(&payload, 60*60*6, token.AccessTokenType)
	if err != nil {
		return nil, err
	}
	r, err := s.tokenBackend.Encode(&payload, 60*60*24*31, token.RefreshTokenType)
	if err != nil {
		return nil, err
	}

	return &Tokens{a, r}, nil
}

var (
	ErrTokenDecoding = errors.New("token decoding error")
)

func (s *authService) Token(
	ctx context.Context,
	accessToken string,
) (*token.User, error) {
	user, err := s.tokenBackend.Decode(accessToken, token.AccessTokenType)
	if err != nil {
		return nil, ErrTokenDecoding
	}

	return token.UserPayloadFromUserClaims(user), nil
}

func (s *authService) RefreshToken(
	ctx context.Context,
	refreshToken string,
) (string, error) {
	refreshTokenClaims, err := s.tokenBackend.Decode(
		refreshToken,
		token.AccessTokenType,
	)
	if err != nil {
		return "", ErrTokenDecoding
	}

	id := refreshTokenClaims.ID
	yes, err := s.userRepo.WasRecentlyBanned(ctx, id)
	if err != nil {
		return "", err
	}

	if yes {
		return "", ErrUserNotActive
	}

	yes, err = s.userRepo.IsKicked(ctx, id)
	if err != nil {
		return "", err
	}

	if yes {
		return "", ErrUserWasKicked
	}

	ml, err := s.userRepo.GetMassLogout(ctx)
	if err != nil {
		return "", err
	}

	if refreshTokenClaims.Iat <= ml.Unix() {
		return "", ErrUserInMassLogout
	}

	user, err := s.userRepo.Get(ctx, refreshTokenClaims.ID)
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

func (s *authService) Me(ctx context.Context, id int32) (*model.Me, error) {
	user, err := s.userRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return model.MeFromUser(user), nil
}
