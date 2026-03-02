package storage

import (
	"sync"
)

var onceUser sync.Once
var userStore UserStore

type UserStore struct {
	users map[string]UserInfo
	mu    sync.RWMutex
}

type UserInfo struct {
	Id       int
	Username string
	Password string
	Email    string
}

func initUserStorage() {
	userStore = UserStore{
		users: make(map[string]UserInfo),
	}
}

func NewUserStore() {
	onceUser.Do(func() {
		initUserStorage()
	})
}

func DoUserWithLock(f func()) {
	userStore.mu.Lock()
	defer userStore.mu.Unlock()
	f()
}

func GetByID(id string) (UserInfo, bool) {
	userStore.mu.RLock()
	defer userStore.mu.RUnlock()
	user, exists := userStore.users[id]
	return user, exists
}

func AddUser(user UserInfo, id string) {
	userStore.mu.Lock()
	defer userStore.mu.Unlock()
	userStore.users[id] = user
}
