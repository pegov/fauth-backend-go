package captcha

type DebugCaptchaClient struct {
	ValidCaptcha string
}

func NewDebugCaptchaClient(validCaptcha string) *DebugCaptchaClient {
	return &DebugCaptchaClient{
		ValidCaptcha: validCaptcha,
	}
}

func (c *DebugCaptchaClient) IsValid(captcha string) bool {
	return c.ValidCaptcha == captcha
}
