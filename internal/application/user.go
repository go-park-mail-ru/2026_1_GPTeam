package application

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

//go:generate mockgen -source=user.go -destination=mocks/user.go -package=mocks
type UserUseCase interface {
	Create(ctx context.Context, user web_helpers.SignupBodyRequest) (web_helpers.AuthUser, error)
	GetById(ctx context.Context, id int) (*models.UserModel, error)
	GetByCredentials(ctx context.Context, user web_helpers.LoginBodyRequest) (*models.UserModel, error)
	IsAuthUserExists(ctx context.Context, isAuth bool, userId int) (web_helpers.User, bool)
	UpdateLastLogin(ctx context.Context, userId int) error
	Update(ctx context.Context, profile models.UpdateUserProfile) (*models.UserModel, error)
	UploadAvatar(ctx context.Context, UserID int, file io.Reader, extension string) (string, error)
	IsStaff(ctx context.Context, userId int) (bool, error)
}
type User struct {
	repository repository.UserRepository
	enumsApp   EnumsUseCase
}

func NewUser(repo repository.UserRepository, enumsApp EnumsUseCase) *User {
	return &User{
		repository: repo,
		enumsApp:   enumsApp,
	}
}

func (obj *User) Create(ctx context.Context, userRequest web_helpers.SignupBodyRequest) (web_helpers.AuthUser, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	bytes, err := bcrypt.GenerateFromPassword([]byte(userRequest.Password), bcrypt.DefaultCost) // ToDo: add pepper (на будущее, так как надо сделать поддержку старых перцов и плавную миграцию на новый перец)
	if err != nil {
		log.Warn("failed to hash password",
			zap.Error(err))
		return web_helpers.AuthUser{}, HashPasswordError
	}
	hashedPassword := string(bytes)
	userModel := models.UserModel{
		Id:        0,
		Username:  userRequest.Username,
		Password:  hashedPassword,
		Email:     userRequest.Email,
		CreatedAt: time.Now(),
		LastLogin: time.Time{},
		AvatarUrl: "",
		UpdatedAt: time.Now(),
		Active:    true,
	}
	id, err := obj.repository.Create(ctx, userModel)
	if err != nil {
		return web_helpers.AuthUser{}, err
	}
	user := web_helpers.AuthUser{
		Id:        id,
		Username:  userModel.Username,
		Email:     userModel.Email,
		LastLogin: userModel.LastLogin,
		CreatedAt: userModel.CreatedAt,
	}
	return user, nil
}

func (obj *User) UploadAvatar(ctx context.Context, userID int, file io.Reader, extension string) (string, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	avatarUrl := uuid.New().String() + extension
	filePath := filepath.Join("./static", avatarUrl)
	dst, err := os.Create(filePath)
	if err != nil {
		log.Warn("failed to create avatar file",
			zap.Int("user_id", userID),
			zap.Error(err))
		return "", err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, file); err != nil {
		log.Warn("failed to copy avatar file",
			zap.Int("user_id", userID),
			zap.Error(err))
		return "", err
	}
	err = obj.repository.UpdateAvatar(ctx, userID, avatarUrl)
	if err != nil {
		return "", err
	}

	return avatarUrl, nil
}

func (obj *User) GetById(ctx context.Context, id int) (*models.UserModel, error) {
	return obj.repository.GetByID(ctx, id)
}

func (obj *User) GetByCredentials(ctx context.Context, user web_helpers.LoginBodyRequest) (*models.UserModel, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	storedUser, err := obj.repository.GetByUsername(ctx, user.Username)
	if err != nil {
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(user.Password))
	if err != nil {
		log.Warn("user not found with credentials",
			zap.Error(err))
		return nil, err
	}
	return storedUser, nil
}

func (obj *User) IsAuthUserExists(ctx context.Context, isAuth bool, userId int) (web_helpers.User, bool) {
	log := logger.GetLoggerWithRequestId(ctx)
	if !isAuth {
		log.Warn("user is not authorized",
			zap.Int("user_id", userId))
		return web_helpers.User{}, false
	}
	storedUser, err := obj.repository.GetByID(ctx, userId)
	if err != nil {
		return web_helpers.User{}, false
	}
	user := web_helpers.User{
		Id:        storedUser.Id,
		Username:  storedUser.Username,
		Email:     storedUser.Email,
		CreatedAt: storedUser.CreatedAt,
		AvatarUrl: storedUser.AvatarUrl,
	}
	return user, true
}

func (obj *User) UpdateLastLogin(ctx context.Context, userId int) error {
	log := logger.GetLoggerWithRequestId(ctx)
	err := obj.repository.UpdateLastLogin(ctx, userId, time.Now())
	if err != nil {
		log.Warn("failed to update last login",
			zap.Int("user_id", userId),
			zap.Error(err))
		return err
	}
	return nil
}

func (obj *User) Update(ctx context.Context, profile models.UpdateUserProfile) (*models.UserModel, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	if profile.Password != nil {
		bytes, err := bcrypt.GenerateFromPassword([]byte(*profile.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Warn("failed to hash password",
				zap.Int("user_id", profile.Id),
				zap.Error(err))
			return nil, HashPasswordError
		}
		hashedPassword := string(bytes)
		profile.Password = &hashedPassword
	}
	return obj.repository.Update(ctx, profile)
}

func (obj *User) IsStaff(ctx context.Context, userId int) (bool, error) {
	user, err := obj.GetById(ctx, userId)
	if err != nil {
		return false, err
	}
	return user.IsStaff, nil
}
