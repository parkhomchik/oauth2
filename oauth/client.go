package oauth

type Client struct {
	ID       string
	Secret   string
	Domain   string
	UserID   string
	Scope    []string
	Password string
}

func (c *Client) GetID() string {
	return c.ID
}

func (c *Client) GetSecret() string {
	return c.Secret
}

func (c *Client) GetDomain() string {
	return c.Domain
}

func (c *Client) GetUserID() string {
	return c.UserID
}

func (c *Client) GetScope() []string {
	return c.Scope
}

func (c *Client) GetPassword() string {
	return c.Password
}
