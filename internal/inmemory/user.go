package inmemory

import "github.com/rdnply/wschat/internal/user"

var _ user.Storage = UserStorage{}

type UserStorage map[string]*user.User

func NewUserStorage() UserStorage {
	return UserStorage(make(map[string]*user.User))
}

func (us UserStorage) Find(login string) bool {
	_, ok := us[login]
	if !ok {
		return false
	}

	return true
}

func (us UserStorage) Add(login string) {
	us[login] = user.New(login)
}
