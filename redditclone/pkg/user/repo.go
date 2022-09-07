package user

import (
	"errors"
	"log"
	"sync"
)

var (
	ErrNoUser        = errors.New("no user found")
	ErrWrongPassword = errors.New("invalid password")
	ErrAlreadyExist  = errors.New("user already exists")
)

type UsersDataRepo struct {
	mu     *sync.RWMutex
	LastID uint32
	Data   map[string]*User
}

func NewUsersRepo() *UsersDataRepo {
	log.Printf("NewUsersRepo: created UsersDataRepo")
	return &UsersDataRepo{
		Data: make(map[string]*User),
		mu:   &sync.RWMutex{},
	}
}

func (ur *UsersDataRepo) Authorize(login, password string) (*User, error) {
	ur.mu.RLock()
	u, ok := ur.Data[login]
	ur.mu.RUnlock()
	if !ok {
		log.Printf("ERROR: Authorize: no user '%v' found", login)
		return nil, ErrNoUser
	}

	if password != u.password {
		log.Printf("ERROR: Authorize: invalid password for user '%v'", login)
		return nil, ErrWrongPassword
	}
	log.Printf("Authorize for '%v'", login)
	return u, nil
}

func (ur *UsersDataRepo) CreateUser(login, pass string) (*User, error) {
	newUser := new(User)
	ur.mu.Lock()
	ur.LastID++
	newUser.ID = ur.LastID
	_, ok := ur.Data[login]

	if ok {
		log.Printf("ERROR: CreateUser, login already exists: '%v'", login)
		return nil, ErrAlreadyExist
	}

	newUser.Username = login
	newUser.password = pass
	ur.Data[login] = newUser
	ur.mu.Unlock()
	log.Printf("CreateUser: created '%v'", login)
	return newUser, nil
}

func (ur *UsersDataRepo) Get(login string) (*User, error) {
	ur.mu.RLock()
	elem, ok := ur.Data[login]
	ur.mu.RUnlock()
	if !ok {
		log.Printf("Get user: no user '%v'", login)
		return nil, ErrNoUser
	}
	return elem, nil
}
