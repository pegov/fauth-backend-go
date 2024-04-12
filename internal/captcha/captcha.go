package captcha

type CaptchaClient interface {
	IsValid(captcha string) bool
}
