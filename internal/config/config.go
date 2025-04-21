package config

type Config struct {
	Database ConfigDatabase
	Cache    ConfigCache
	HTTP     ConfigHTTP
	SMTP     ConfigSMTP
	Captcha  ConfigCaptcha
	OAuth    ConfigOAuth
	App      ConfigApp
	Flags    Flags
}

type ConfigDatabase struct {
	URL             string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime int
}

type ConfigCache struct {
	URL string
}

type ConfigHTTP struct {
	Domain string
	Secure bool
}

type ConfigSMTP struct {
	Username string
	Password string
	Host     string
	Port     string
}

type ConfigCaptcha struct {
	RecaptchaSecret string
}

type ConfigOAuth struct {
	Providers          []string
	GoogleClientID     string
	GoogleClientSecret string
	VKAppID            string
	VKAppSecret        string
}

type ConfigApp struct {
	LoginRatelimit         int
	AccessTokenCookieName  string
	RefreshTokenCookieName string
	AcessTokenExpiration   int
	RefreshTokenExpiration int
}

type Flags struct {
	Host                                  string
	Port                                  int
	Debug, Verbose, Test                  bool
	AccessLog, ErrorLog                   string
	PrivateKeyPath, PublicKeyPath, JWTKID string
}
