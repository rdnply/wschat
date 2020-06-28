package user

type User struct {
	login string
}

func New(login string) *User {
	return &User{
		login: login,
	}
}

type Storage interface {
	Find(login string) bool
	Add(login string)
}
