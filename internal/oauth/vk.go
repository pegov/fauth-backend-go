package oauth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type VKOAuthProvider struct {
	ClientID     string
	ClientSecret string
	isLoginOnly  bool
}

type VKOAuthResponse struct {
	Email  *string `json:"email"`
	UserID int64   `json:"user_id"`
}

func NewVKOAuthProvider(clientID, clientSecret string, isLoginOnly bool) VKOAuthProvider {
	return VKOAuthProvider{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		isLoginOnly:  isLoginOnly,
	}
}

func (p *VKOAuthProvider) Name() string {
	return "vk"
}

func (p *VKOAuthProvider) IsLoginOnly() bool {
	return p.isLoginOnly
}

func (p *VKOAuthProvider) CreateOAuthURI(redirectURI, state string) string {
	return fmt.Sprintf(
		"https://oauth.vk.com/authorize"+
			"?client_id=%s"+
			"&scope=email"+
			"&redirect_uri=%s"+
			"&response_type=code"+
			"&v=5.122"+
			"&state=%s",
		p.ClientID,
		redirectURI,
		state,
	)
}

func (p *VKOAuthProvider) GetUserData(redirectURI, code string) (*OAuthResponse, error) {
	req, err := http.NewRequest("GET", "https://oauth.vk.com/access_token", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("client_id", p.ClientID)
	q.Add("client_secret", p.ClientSecret)
	q.Add("code", code)
	q.Add("redirect_uri", redirectURI)
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

	var response VKOAuthResponse
	if err = json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	if response.Email == nil || strings.TrimSpace(*response.Email) == "" {
		return nil, errors.New("email not found")
	}

	return &OAuthResponse{
		SID:   strconv.FormatInt(response.UserID, 10),
		Email: *response.Email,
	}, nil
}
