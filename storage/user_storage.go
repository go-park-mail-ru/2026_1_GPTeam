package storage

import (
	"main/base"
	"sync"
	"time"
)

var onceUser sync.Once
var userStore UserStore

type UserStore struct {
	users map[int]UserInfo
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
	Balance   int
}

func initUserStorage() {
	userStore = UserStore{
		users: make(map[int]UserInfo),
	}
}

func NewUserStore() {
	onceUser.Do(func() {
		initUserStorage()
	})
}

func GetUserByID(id int) (UserInfo, bool) {
	userStore.mu.RLock()
	defer userStore.mu.RUnlock()
	user, exists := userStore.users[id]
	return user, exists
}

func AddUser(user UserInfo) int {
	userStore.mu.Lock()
	defer userStore.mu.Unlock()
	id := len(userStore.users)
	user.Id = id
	userStore.users[id] = user
	return id
}

func FindUserByCredentials(user base.LoginBodyRequest) (UserInfo, bool) {
	for _, value := range userStore.users {
		if value.Username == user.Username && value.Password == user.Password {
			return value, true
		}
	}
	return UserInfo{}, false
}
