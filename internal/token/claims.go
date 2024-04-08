package token

import "github.com/golang-jwt/jwt/v5"

const (
	AccessTokenType  = "access"
	RefreshTokenType = "refresh"
)

type TokenHeader struct {
	Kid string `json:"kid"`
}

type UserClaims struct {
	Iat      int64    `json:"iat"`
	Exp      int64    `json:"exp"`
	Type     string   `json:"type"`
	Id       int32    `json:"id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

type User struct {
	Id       int32    `json:"id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
}

func UserPayloadFromUserClaims(p *UserClaims) *User {
	return &User{
		Id:       p.Id,
		Username: p.Username,
		Roles:    p.Roles,
	}
}
