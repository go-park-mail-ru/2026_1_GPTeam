package storage

import (
	"main/base"
	"sync"
	"testing"
	"time"
)

func setupUserStoreTest() {
	onceUser = sync.Once{}
	NewUserStore()
}

func TestAddUserAndGetUserByID(t *testing.T) {
	setupUserStoreTest()

	createdAt := time.Now().UTC().Truncate(time.Second)
	id := AddUser(UserInfo{
		Username:  "WatchDemo",
		Password:  "secret",
		Email:     "WatchDemo@gmail.com",
		CreatedAt: createdAt,
	})

	if id != 0 {
		t.Fatalf("expected first user id 0, got %d", id)
	}

	user, ok := GetUserByID(id)
	if !ok {
		t.Fatal("expected user to exist")
	}
	if user.Username != "WatchDemo" || user.Email != "WatchDemo@gmail.com" {
		t.Fatalf("unexpected user data: %+v", user)
	}
}

func TestFindUserByCredentials(t *testing.T) {
	setupUserStoreTest()

	AddUser(UserInfo{Username: "Grisha", Password: "123456", Email: "Grisha@example.com"})

	user, ok := FindUserByCredentials(base.LoginBodyRequest{Username: "Grisha", Password: "123456"})
	if !ok {
		t.Fatal("expected credentials to be found")
	}
	if user.Username != "Grisha" {
		t.Fatalf("expected Grisha, got %q", user.Username)
	}

	_, ok = FindUserByCredentials(base.LoginBodyRequest{Username: "Grisha", Password: "wrong"})
	if ok {
		t.Fatal("expected wrong password to fail")
	}
}

func TestUserExistsAndEmailExists(t *testing.T) {
	setupUserStoreTest()

	AddUser(UserInfo{Username: "Timur", Password: "pw", Email: "Timur@example.com"})

	if !UserExists("Timur") {
		t.Fatal("expected username to exist")
	}
	if UserExists("nobody") {
		t.Fatal("did not expect username to exist")
	}
	if !EmailExists("Timur@example.com") {
		t.Fatal("expected email to exist")
	}
	if EmailExists("none@example.com") {
		t.Fatal("did not expect email to exist")
	}
}

func TestIsAuthUserInDatabase(t *testing.T) {
	setupUserStoreTest()

	id := AddUser(UserInfo{
		Username:        "kek",
		Password:        "pw",
		Email:           "kek@example.com",
		Balance:         150.5,
		BalanceCurrency: "RUB",
		AvatarUrl:       "/avatar.png",
		CreatedAt:       time.Unix(1000, 0),
		LastLogin:       time.Unix(2000, 0),
	})

	authUser, ok := IsAuthUserInDatabase(true, "0")
	if !ok {
		t.Fatal("expected authenticated user to be returned")
	}
	if authUser.Username != "kek" || authUser.Email != "kek@example.com" {
		t.Fatalf("unexpected auth user data: %+v", authUser)
	}
	if authUser.Balance != 150.5 || authUser.BalanceCurrency != "RUB" {
		t.Fatalf("unexpected balance info: %+v", authUser)
	}
	if id != 0 {
		t.Fatalf("expected id 0, got %d", id)
	}

	_, ok = IsAuthUserInDatabase(false, "0")
	if ok {
		t.Fatal("expected false when isAuth is false")
	}

	_, ok = IsAuthUserInDatabase(true, "bad-id")
	if ok {
		t.Fatal("expected false for invalid user id")
	}

	_, ok = IsAuthUserInDatabase(true, "10")
	if ok {
		t.Fatal("expected false for missing user")
	}
}
