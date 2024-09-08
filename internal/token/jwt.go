package token

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JwtBackend interface {
	Encode(payload *User, exp time.Duration, tokenType string) (string, error)
	Decode(token string, tokenType string) (*UserClaims, error)
}

type jwtBackend struct {
	PrivateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey
	CurrentKid string
}

func NewJwtBackend(privateKeyBytes []byte, publicKeyBytes []byte, currentKid string) JwtBackend {
	privateBlock, _ := pem.Decode(privateKeyBytes)
	privateParsed, err := x509.ParsePKCS8PrivateKey(privateBlock.Bytes)
	if err != nil {
		log.Fatalln(err)
	}
	privateKey := privateParsed.(ed25519.PrivateKey)

	publicBlock, _ := pem.Decode(publicKeyBytes)
	publicParsed, _ := x509.ParsePKIXPublicKey(publicBlock.Bytes)
	if err != nil {
		log.Fatalln(err)
	}
	publicKey := publicParsed.(ed25519.PublicKey)

	return &jwtBackend{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		CurrentKid: currentKid,
	}
}

func NewJwtBackendRaw(privateKeyBytesRaw, publicKeyBytesRaw []byte, currentKid string) JwtBackend {
	return &jwtBackend{
		PrivateKey: privateKeyBytesRaw,
		PublicKey:  publicKeyBytesRaw,
		CurrentKid: currentKid,
	}
}

func (backend *jwtBackend) Encode(payload *User, expiration time.Duration, tokenType string) (string, error) {
	token := jwt.New(jwt.SigningMethodEdDSA)
	token.Header["kid"] = backend.CurrentKid

	claims := UserClaims{
		ID:       payload.ID,
		Username: payload.Username,
		Roles:    payload.Roles,
	}
	iat := time.Now()
	exp := iat.Add(expiration)
	claims.IssuedAt = jwt.NewNumericDate(iat)
	claims.ExpiresAt = jwt.NewNumericDate(exp)
	claims.Type = tokenType

	token.Claims = claims

	tokenString, err := token.SignedString(backend.PrivateKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

var (
	ErrJwtDecodeMissingKid       = errors.New("jwt decode missing kid")
	ErrJwtDecodeInvalidKid       = errors.New("jwt decode invalid kid")
	ErrJwtDecodeInvalidToken     = errors.New("jwt decode invalid token")
	ErrJwtDecodeInvalidTokenType = errors.New("jwt decode invalid token type")
)

func (backend *jwtBackend) Decode(tokenString string, tokenType string) (*UserClaims, error) {
	var payload UserClaims
	token, err := jwt.ParseWithClaims(tokenString, &payload, func(t *jwt.Token) (any, error) {
		kid, ok := t.Header["kid"]
		if !ok {
			return nil, ErrJwtDecodeMissingKid
		}

		if kid != backend.CurrentKid {
			return nil, ErrJwtDecodeInvalidKid
		}

		return backend.PublicKey, nil
	})

	if err != nil || !token.Valid {
		return nil, ErrJwtDecodeInvalidToken
	}

	if payload.Type != tokenType {
		return nil, ErrJwtDecodeInvalidTokenType
	}

	return &payload, nil
}
