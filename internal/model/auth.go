package model

import (
	"errors"
	"regexp"
	"slices"
	"strings"

	"github.com/pegov/fauth-backend-go/internal/entity"
)

type RegisterRequest struct {
	Email     string `json:"email"`
	Username  string `json:"username"`
	Password1 string `json:"password1"`
	Password2 string `json:"password2"`
	Captcha   string `json:"captcha"`
}

func (r *RegisterRequest) Validate() error {
	email, err := validateEmail(r.Email)
	if err != nil {
		return NewValidationError(err)
	}
	r.Email = email

	username, err := validateUsername(r.Username)
	if err != nil {
		return NewValidationError(err)
	}
	r.Username = username

	if err := validatePassword(r.Password1, r.Password2); err != nil {
		return NewValidationError(err)
	}

	return nil
}

var SimpleEmailRe = regexp.MustCompile(`^[\w\-\.]+@([\w-]+\.)+[\w-]{2,6}$`)

var (
	ErrEmailEmpty            = errors.New("email empty")
	ErrEmailMissingSeparator = errors.New("email no @")
	ErrEmailLength           = errors.New("email length")
	ErrEmailWrong            = errors.New("email wrong")
)

var (
	maxUserPartLen = 64
	maxDomainPart  = 255
)

func validateEmail(email string) (string, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return "", ErrEmailEmpty
	}

	if !strings.Contains(email, "@") {
		return "", ErrEmailMissingSeparator
	}

	parts := strings.SplitN(email, "@", 2)

	userPart := parts[0]
	domainPart := parts[1]

	if len(userPart) > maxUserPartLen || len(domainPart) > maxDomainPart {
		return "", ErrEmailLength
	}

	if !SimpleEmailRe.MatchString(email) {
		return "", ErrEmailWrong
	}

	return email, nil
}

var (
	ErrUsernameForbidden         = errors.New("username forbidden")
	ErrUsernameForbiddenChar     = errors.New("username chars")
	ErrUsernameLength            = errors.New("username length")
	ErrUsernameMustContainLetter = errors.New("username must contain letter")
	ErrUsernameDifferentLetters  = errors.New("username different letters")
)

var (
	forbiddenUsernames = []string{"admin", "moderator"}
)

var (
	usernameAllowedCharsEn = "abcdefghijklmopqrstuvwxyzABCDEFGHIJKLMOPQRSTUVWXYZ"
	usernameAllowedCharsRu = "абвгдеёжзийклмнопрстуфхцчшщъыьэюяАБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЪЫЬЭЮЯ"
)

const (
	usernameMinLength = 4
	usernameMaxLength = 20
)

func validateUsername(username string) (string, error) {
	username = strings.TrimSpace(username)

	if slices.Contains(forbiddenUsernames, username) {
		return "", ErrUsernameForbidden
	}

	var hasEn, hasRu bool

	for _, usernameChar := range username {
		ok := false
		for _, allowedChar := range usernameAllowedCharsEn {
			if usernameChar == allowedChar {
				hasEn = true
				ok = true
				break
			}
		}

		for _, allowedChar := range usernameAllowedCharsRu {
			if usernameChar == allowedChar {
				hasRu = true
				ok = true
				break
			}
		}

		if !ok {
			return "", ErrUsernameForbiddenChar
		}
	}

	if hasEn && hasRu {
		return "", ErrUsernameDifferentLetters
	}

	if len(username) < usernameMinLength || len(username) > usernameMaxLength {
		return "", ErrUsernameLength
	}
	return username, nil
}

var (
	ErrPasswordMismatch      = errors.New("password mismatch")
	ErrPasswordLength        = errors.New("password length")
	ErrPasswordForbiddenChar = errors.New("password chars")
)

var (
	passwordAllowedChars = "abcdefghijklmopqrstuvwxyzABCDEFGHIJKLMOPQRSTUVWXYZ123456 !\"#$%&'()*+,-./:;<=>?@[]^_`{|}~"
)

const (
	passwordMinLength = 6
	passwordMaxLength = 32
)

func validatePassword(password1 string, password2 string) error {
	if password1 != password2 {
		return ErrPasswordMismatch
	}

	if len(password1) < passwordMinLength || len(password1) > passwordMaxLength {
		return ErrPasswordLength
	}

	for _, passwordChar := range password1 {
		ok := false
		for _, allowedChar := range passwordAllowedChars {
			if passwordChar == allowedChar {
				ok = true
				break
			}
		}

		if !ok {
			return ErrPasswordForbiddenChar
		}
	}

	return nil
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Me struct {
	ID       int32  `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Verified bool   `json:"verified"`
	// Roles
	// OAuth
}

func MeFromUser(user *entity.User) *Me {
	return &Me{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
		Verified: user.Verified,
	}
}
