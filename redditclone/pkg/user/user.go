package user

type User struct {
	ID       uint32 `json:"id"`
	Username string `json:"username"`
	password string
}

type UsersRepo interface {
	Authorize(login, pass string) (*User, error)
	CreateUser(login, pass string) (*User, error)
	Get(login string) (*User, error)
}
