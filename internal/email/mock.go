package email

type MockEmailClient struct{}

func NewMockEmailClient() EmailClient {
	return &MockEmailClient{}
}

func (c *MockEmailClient) SendEmail(from, to, message string) error {
	return nil
}
