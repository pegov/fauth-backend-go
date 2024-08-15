package handler

import (
	"net/http"

	"github.com/pegov/fauth-backend-go/internal/service"
)

type OAuthHandler interface {
	Login(w http.ResponseWriter, r *http.Request) error
	Callback(w http.ResponseWriter, r *http.Request) error
}

type oauthHandler struct {
	oauthService service.OAuthService
}

func NewOAuthHandler(oauthService service.OAuthService) *oauthHandler {
	return &oauthHandler{
		oauthService: oauthService,
	}
}
