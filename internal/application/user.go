package application

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type UserUseCase interface {
	Create(ctx context.Context, user web_helpers.SignupBodyRequest) (web_helpers.AuthUser, error)
	GetById(ctx context.Context, id int) (*models.UserModel, error)
	GetByCredentials(ctx context.Context, user web_helpers.LoginBodyRequest) (*models.UserModel, error)
	IsAuthUserExists(ctx context.Context, isAuth bool, userId int) (web_helpers.User, bool)
	UpdateLastLogin(ctx context.Context, userId int) error
	Update(ctx context.Context, profile models.UpdateUserProfile) (*models.UserModel, error)
}

type User struct {
	repository repository.UserRepository
	log        *zap.Logger
}

func NewUser(repository repository.UserRepository) *User {
	return &User{
		repository: repository,
		log:        logger.GetLogger(),
	}
}

func (obj *User) Create(ctx context.Context, userRequest web_helpers.SignupBodyRequest) (web_helpers.AuthUser, error) {
	obj.log.Info("creating user",
		zap.String("username", userRequest.Username),
		zap.String("request_id", ctx.Value("request_id").(string)))
	bytes, err := bcrypt.GenerateFromPassword([]byte(userRequest.Password), bcrypt.DefaultCost)
	if err != nil {
		obj.log.Warn("failed to hash password",
			zap.String("request_id", ctx.Value("request_id").(string)),
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

func (obj *User) GetById(ctx context.Context, id int) (*models.UserModel, error) {
	obj.log.Info("getting user by id",
		zap.Int("id", id),
		zap.String("request_id", ctx.Value("request_id").(string)))
	return obj.repository.GetByID(ctx, id)
}

func (obj *User) GetByCredentials(ctx context.Context, user web_helpers.LoginBodyRequest) (*models.UserModel, error) {
	obj.log.Info("getting user by credentials",
		zap.String("username", user.Username),
		zap.String("request_id", ctx.Value("request_id").(string)))
	storedUser, err := obj.repository.GetByUsername(ctx, user.Username)
	if err != nil {
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(user.Password))
	if err != nil {
		obj.log.Warn("user not found",
			zap.String("request_id", ctx.Value("request_id").(string)),
			zap.Error(err))
		return nil, err
	}
	return storedUser, nil
}

func (obj *User) IsAuthUserExists(ctx context.Context, isAuth bool, userId int) (web_helpers.User, bool) {
	obj.log.Info("checking user by id",
		zap.Int("user_id", userId),
		zap.String("request_id", ctx.Value("request_id").(string)))
	if !isAuth {
		obj.log.Warn("user is not authorized",
			zap.Int("user_id", userId),
			zap.String("request_id", ctx.Value("request_id").(string)))
		return web_helpers.User{}, false
	}
	storedUser, err := obj.repository.GetByID(ctx, userId)
	if err != nil {
		obj.log.Warn("user not found in db",
			zap.Int("user_id", userId),
			zap.String("request_id", ctx.Value("request_id").(string)),
			zap.Error(err))
		return web_helpers.User{}, false
	}
	user := web_helpers.User{
		Username:  storedUser.Username,
		Email:     storedUser.Email,
		CreatedAt: storedUser.CreatedAt,
		AvatarUrl: storedUser.AvatarUrl,
	}
	return user, true
}

func (obj *User) UpdateLastLogin(ctx context.Context, userId int) error {
	obj.log.Info("updating last login field",
		zap.Int("user_id", userId),
		zap.String("request_id", ctx.Value("request_id").(string)))
	err := obj.repository.UpdateLastLogin(ctx, userId, time.Now())
	if err != nil {
		return err
	}
	return nil
}

func (obj *User) Update(ctx context.Context, profile models.UpdateUserProfile) (*models.UserModel, error) {
	obj.log.Info("updating profile",
		zap.Int("user_id", profile.Id),
		zap.String("request_id", ctx.Value("request_id").(string)))
	if profile.Password != nil {
		bytes, err := bcrypt.GenerateFromPassword([]byte(*profile.Password), bcrypt.DefaultCost)
		if err != nil {
			obj.log.Warn("failed to hash password",
				zap.Int("user_id", profile.Id),
				zap.String("request_id", ctx.Value("request_id").(string)),
				zap.Error(err))
			return nil, HashPasswordError
		}
		hashedPassword := string(bytes)
		profile.Password = &hashedPassword
	}
	return obj.repository.Update(ctx, profile)
}
