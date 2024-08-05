package oauth

type OAuthResponse struct {
	SID   string
	Email string
}

type OAuthProvider interface {
	Name() string
	CreateOAuthURI(redirectURI, state string) string
	IsLoginOnly() bool
	GetUserData(redirectURI, code string) (*OAuthResponse, error)
}
