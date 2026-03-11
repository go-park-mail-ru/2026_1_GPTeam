package storage_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-park-mail-ru/2026_1_GPTeam/base"
	"github.com/go-park-mail-ru/2026_1_GPTeam/storage"
)

func setupUserStoreTest(t *testing.T) {
	t.Helper()
	storage.NewUserStore()
}

func TestAddUserAndGetUserByID(t *testing.T) {
	setupUserStoreTest(t)

	createdAt := time.Now().UTC().Truncate(time.Second)
	id := storage.AddUser(storage.UserInfo{
		Username:  "WatchDemo_user_storage_test",
		Password:  "secret",
		Email:     "watchdemo_user_storage_test@gmail.com",
		CreatedAt: createdAt,
	})

	user, ok := storage.GetUserByID(id)
	require.True(t, ok)

	assert.Equal(t, "WatchDemo_user_storage_test", user.Username)
	assert.Equal(t, "secret", user.Password)
	assert.Equal(t, "watchdemo_user_storage_test@gmail.com", user.Email)
	assert.Equal(t, createdAt, user.CreatedAt)
}

func TestFindUserByCredentialsReturnsUser(t *testing.T) {
	setupUserStoreTest(t)

	createdAt := time.Now().UTC().Truncate(time.Second)
	storage.AddUser(storage.UserInfo{
		Username:  "CredUser_user_storage_test",
		Password:  "verysecret",
		Email:     "cred_user_storage_test@gmail.com",
		CreatedAt: createdAt,
	})

	user, ok := storage.FindUserByCredentials(base.LoginBodyRequest{
		Username: "CredUser_user_storage_test",
		Password: "verysecret",
	})
	require.True(t, ok)

	assert.Equal(t, "CredUser_user_storage_test", user.Username)
	assert.Equal(t, "verysecret", user.Password)
	assert.Equal(t, "cred_user_storage_test@gmail.com", user.Email)
}

func TestFindUserByCredentialsWrongPassword(t *testing.T) {
	setupUserStoreTest(t)

	createdAt := time.Now().UTC().Truncate(time.Second)
	storage.AddUser(storage.UserInfo{
		Username:  "WrongPassword_user_storage_test",
		Password:  "correct-password",
		Email:     "wrong_password_user_storage_test@gmail.com",
		CreatedAt: createdAt,
	})

	_, ok := storage.FindUserByCredentials(base.LoginBodyRequest{
		Username: "WrongPassword_user_storage_test",
		Password: "incorrect-password",
	})
	assert.False(t, ok)
}

func TestUserExistsReturnsTrueForExistingUser(t *testing.T) {
	setupUserStoreTest(t)

	createdAt := time.Now().UTC().Truncate(time.Second)
	storage.AddUser(storage.UserInfo{
		Username:  "ExistsUser_user_storage_test",
		Password:  "secret",
		Email:     "exists_user_storage_test@gmail.com",
		CreatedAt: createdAt,
	})

	assert.True(t, storage.UserExists("ExistsUser_user_storage_test"))
	assert.False(t, storage.UserExists("DefinitelyMissingUser_user_storage_test"))
}

func TestEmailExistsReturnsTrueForExistingEmail(t *testing.T) {
	setupUserStoreTest(t)

	createdAt := time.Now().UTC().Truncate(time.Second)
	storage.AddUser(storage.UserInfo{
		Username:  "EmailUser_user_storage_test",
		Password:  "secret",
		Email:     "email_user_storage_test@gmail.com",
		CreatedAt: createdAt,
	})

	assert.True(t, storage.EmailExists("email_user_storage_test@gmail.com"))
	assert.False(t, storage.EmailExists("missing_email_user_storage_test@gmail.com"))
}

func TestIsAuthUserInDatabaseReturnsTrueForExistingID(t *testing.T) {
	setupUserStoreTest(t)

	createdAt := time.Now().UTC().Truncate(time.Second)
	id := storage.AddUser(storage.UserInfo{
		Username:  "AuthUser_user_storage_test",
		Password:  "secret",
		Email:     "auth_user_storage_test@gmail.com",
		CreatedAt: createdAt,
		LastLogin: createdAt,
		AvatarUrl: "avatar.png",
		Balance:   150.5,
	})

	user, ok := storage.IsAuthUserInDatabase(true, "0")
	if id != 0 {
		user, ok = storage.IsAuthUserInDatabase(true, string(rune('0'+id)))
	}

	require.True(t, ok)
	assert.Equal(t, "AuthUser_user_storage_test", user.Username)
	assert.Equal(t, "auth_user_storage_test@gmail.com", user.Email)
	assert.Equal(t, "avatar.png", user.AvatarUrl)
	assert.Equal(t, 150.5, user.Balance)
}

func TestIsAuthUserInDatabaseReturnsFalseForInvalidAuthData(t *testing.T) {
	setupUserStoreTest(t)

	_, ok := storage.IsAuthUserInDatabase(false, "1")
	assert.False(t, ok)

	_, ok = storage.IsAuthUserInDatabase(true, "invalid-id")
	assert.False(t, ok)

	_, ok = storage.IsAuthUserInDatabase(true, "999999")
	assert.False(t, ok)
}
