package service_test

import (
	"crypto/ed25519"
	"strings"
	"testing"
	"time"

	"github.com/ovechkin-dm/mockio/v2/mock"
	"github.com/stretchr/testify/require"

	"github.com/pegov/fauth-backend-go/internal/captcha"
	"github.com/pegov/fauth-backend-go/internal/email"
	"github.com/pegov/fauth-backend-go/internal/entity"
	"github.com/pegov/fauth-backend-go/internal/model"
	"github.com/pegov/fauth-backend-go/internal/password"
	"github.com/pegov/fauth-backend-go/internal/repo"
	"github.com/pegov/fauth-backend-go/internal/service"
	"github.com/pegov/fauth-backend-go/internal/token"
)

func TestLogin(t *testing.T) {
	ctrl := mock.NewMockController(t)
	repoM := mock.Mock[repo.UserRepo](ctrl)
	ph := password.NewPlainTextPasswordHasher()

	req := model.LoginRequest{
		Login:    "user",
		Password: "pass",
	}
	res := entity.User{
		ID:        1,
		Email:     "",
		Username:  req.Login,
		Password:  &req.Password,
		Active:    true,
		Verified:  true,
		CreatedAt: time.Time{},
		LastLogin: time.Time{},
	}

	mock.WhenDouble(repoM.GetByLogin(mock.AnyContext(), mock.Exact(req.Login))).ThenReturn(&res, nil)

	generateKeys := func(seed []byte) ([]byte, []byte) {
		private := ed25519.NewKeyFromSeed(seed)
		public := private.Public().(ed25519.PublicKey)
		return private, public
	}
	privateKey, publicKey := generateKeys([]byte(strings.Repeat("a", ed25519.SeedSize)))
	tokenBackend := token.NewJwtBackendRaw(privateKey, publicKey, "1")

	s := service.NewAuthService(
		repoM,
		captcha.NewDebugCaptchaClient(""),
		ph,
		tokenBackend,
		email.NewMockEmailClient(),
	)

	tokens, err := s.Login(t.Context(), &req)
	require.NoError(t, err)

	require.NotEmpty(t, tokens.Access)
	require.NotEmpty(t, tokens.Refresh)
}
