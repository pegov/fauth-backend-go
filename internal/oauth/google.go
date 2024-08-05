package oauth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

type GoogleOAuthProvider struct {
	ClientID     string
	ClientSecret string
	isLoginOnly  bool
}

type GoogleOAuthResponse struct {
	IDToken string `json:"id_token"`
}

func NewGoogleOAuthProvider(clientID, clientSecret string, isLoginOnly bool) GoogleOAuthProvider {
	return GoogleOAuthProvider{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		isLoginOnly:  isLoginOnly,
	}
}

func (p *GoogleOAuthProvider) Name() string {
	return "google"
}

func (p *GoogleOAuthProvider) IsLoginOnly() bool {
	return p.isLoginOnly
}

func (p *GoogleOAuthProvider) CreateOAuthURI(redirectURI, state string) string {
	return fmt.Sprintf(
		"https://accounts.google.com/o/oauth2/v2/auth"+
			"?scope=email%%20profile"+
			"&response_type=code"+
			"&state=%s"+
			"redirect_uri=%s"+
			"&client_id=%s",
		state,
		redirectURI,
		p.ClientID,
	)
}

func (p *GoogleOAuthProvider) GetUserData(redirectURI, code string) (*OAuthResponse, error) {
	req, err := http.NewRequest("POST", "https://oauth2.googleapis.com/token", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("content-length", "0")

	q := req.URL.Query()
	q.Add("client_id", p.ClientID)
	q.Add("client_secret", p.ClientSecret)
	q.Add("code", code)
	q.Add("redirect_uri", redirectURI)
	q.Add("grant_type", "authorization_code")
	req.URL.RawQuery = q.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, errors.New("status != 200")
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var response GoogleOAuthResponse
	if err = json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	token, err := jwt.Parse(response.IDToken, nil, jwt.WithoutClaimsValidation())
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("failed to cast response to map")
	}

	sub, ok := claims["sub"]
	if !ok {
		return nil, errors.New("failed to get \"sub\"")
	}
	email, ok := claims["email"]
	if !ok {
		return nil, errors.New("failed to get \"email\"")
	}

	return &OAuthResponse{
		SID:   sub.(string),
		Email: email.(string),
	}, nil
}
