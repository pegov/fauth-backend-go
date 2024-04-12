package captcha

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ReCaptchaClient struct {
	secret string
}

func NewReCaptchaClient(secret string) CaptchaClient {
	return &ReCaptchaClient{
		secret: secret,
	}
}

type response struct {
	Success bool `json:"success"`
	// ChallengeTS string   `json:"challenge_ts"`
	// Hostname    string   `json:"hostname"`
	// ErrorCodes  []string `json:"error-codes"`
}

func (c *ReCaptchaClient) IsValid(captcha string) bool {
	res, err := sendReCaptchaRequest(c.secret, captcha)
	if err != nil {
		// TODO: log
		return false
	}

	return res.Success
}

func sendReCaptchaRequest(secret string, captcha string) (*response, error) {
	data := map[string]string{
		"secret":   secret,
		"response": captcha,
	}
	jsonValue, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON data: %w", err)
	}

	recaptchaURL := "https://www.google.com/recaptcha/api/siteverify"
	res, err := http.Post(recaptchaURL, "application/x-www-form-urlencoded", bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var recaptchaResp response
	err = json.Unmarshal(body, &recaptchaResp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON response: %w", err)
	}

	return &recaptchaResp, nil
}
