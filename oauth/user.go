package oauth

type User struct {
	UserName string
	Password string
	Scope    []string
}

func (u *User) GetUserName() string {
	return u.UserName
}

func (u *User) GetPassword() string {
	return u.Password
}

func (u *User) GetScope() []string {
	return u.Scope
}
