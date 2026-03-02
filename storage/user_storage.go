package storage

import (
	"main/base"
	"strconv"
	"sync"
	"time"
)

var onceUser sync.Once
var userStore UserStore

type UserStore struct {
	users map[string]UserInfo
	mu    sync.RWMutex
}

type UserInfo struct {
	Id        int
	Username  string
	Password  string
	Email     string
	CreatedAt time.Time
	LastLogin time.Time
	AvatarUrl string
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

func GetUserByID(id string) (UserInfo, bool) {
	userStore.mu.RLock()
	defer userStore.mu.RUnlock()
	user, exists := userStore.users[id]
	return user, exists
}

func AddUser(user UserInfo) string {
	userStore.mu.Lock()
	defer userStore.mu.Unlock()
	id := len(userStore.users)
	user.Id = id
	key := strconv.Itoa(id)
	userStore.users[key] = user
	return key
}

func FindUserByCredentials(user base.LoginBodyRequest) (UserInfo, bool) {
	for _, value := range userStore.users {
		if value.Username == user.Username && value.Password == user.Password {
			return value, true
		}
	}
	return UserInfo{}, false
}
