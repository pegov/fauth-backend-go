package service

import (
	"context"

	"github.com/pegov/fauth-backend-go/internal/model"
	"github.com/pegov/fauth-backend-go/internal/repo"
	"github.com/pegov/fauth-backend-go/internal/token"
)

type OAuthService interface {
	Login(ctx context.Context, request *model.LoginRequest) (string, string, error)
	Callback(ctx context.Context, accessToken string) (*token.User, error)
}

type oauthService struct {
	userRepo     repo.UserRepo
	tokenBackend token.JwtBackend
}

func NewOAuthService(
	userRepo repo.UserRepo,
	tokenBackend token.JwtBackend,
) *oauthService {
	return &oauthService{
		userRepo:     userRepo,
		tokenBackend: tokenBackend,
	}
}
