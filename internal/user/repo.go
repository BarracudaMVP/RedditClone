package user

import (
	"RedditClone/internal/id"
	"errors"
	"strings"
	"sync"
)

var ErrUserExists = errors.New("user with current username exists")
var ErrUserNotFound = errors.New("user not found")

type UserRepository interface {
	Create(username Username, passwordHash string) (User, error)
	GetByUsername(username Username) (StoredUser, error)
}
type MemoryUserRepository struct {
	storedUsers []StoredUser
	mu          sync.RWMutex
}

func NewMemoryUserRepository() *MemoryUserRepository {
	return &MemoryUserRepository{storedUsers: []StoredUser{
		{
			User: User{
				ID:       "507f1f77bcf86cd799439011",
				Username: "alice",
				Admin:    false,
			},
			//password: alice12345
			PasswordHash: "$2a$10$HJR14d5BhJZqA2OE3cYfY.lpyu2MSwAjKHXMHuIORZNG63zEzZ01u",
		},
	},
	}
}

func (repo *MemoryUserRepository) existsByID(id id.ID) bool {
	for i := range repo.storedUsers {
		if id == repo.storedUsers[i].User.ID {
			return true
		}
	}
	return false
}
func (repo *MemoryUserRepository) Create(username Username, passwordHash string) (User, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	err := ValidateUsername(username)
	if err != nil {
		return User{}, err
	}
	for i := range repo.storedUsers {
		if strings.EqualFold(string(username), string(repo.storedUsers[i].User.Username)) {
			return User{}, ErrUserExists
		}
	}
	var uid id.ID
	for {
		uid, err = id.NewID()
		if err != nil {
			return User{}, err
		}
		if !repo.existsByID(uid) {
			break
		}
	}
	u := User{
		ID:       uid,
		Username: username,
		Admin:    false,
	}
	s := StoredUser{
		User:         u,
		PasswordHash: passwordHash,
	}
	repo.storedUsers = append(repo.storedUsers, s)
	return u, nil
}

func (repo *MemoryUserRepository) GetByUsername(username Username) (StoredUser, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	for i := range repo.storedUsers {
		if strings.EqualFold(string(repo.storedUsers[i].User.Username), string(username)) {
			return repo.storedUsers[i], nil
		}
	}
	return StoredUser{}, ErrUserNotFound
}
