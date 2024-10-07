package email

import (
	"fmt"
	"net/smtp"
)

type EmailClient interface {
	SendEmail(from, to, message string) error
}

type StdEmailClient struct {
	username string
	password string
	host     string
	port     string
}

func NewStdEmailClient(username, password, host, port string) EmailClient {
	return &StdEmailClient{
		username: username,
		password: password,
		host:     host,
		port:     port,
	}
}

func (c *StdEmailClient) SendEmail(from, to, message string) error {
	auth := smtp.PlainAuth("", c.username, c.password, c.host)
	addr := fmt.Sprintf("%s:%s", c.host, c.port)
	if err := smtp.SendMail(
		addr,
		auth,
		from,
		[]string{to},
		[]byte(message),
	); err != nil {
		return err
	}

	return nil
}
