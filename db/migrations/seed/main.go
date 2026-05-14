package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func addUser(conn *pgx.Conn, username string, plainPassword string, email string, isStaff bool) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	password := string(bytes)
	query := `insert into "user" (username, password, email, is_staff) VALUES ($1, $2, $3, $4);`
	_, err = conn.Exec(context.Background(), query, username, password, email, isStaff)
	if err != nil {
		pgErr, ok := errors.AsType[*pgconn.PgError](err)
		if ok {
			if pgErr.Code != "23505" {
				panic(err)
			}
		} else {
			panic(err)
		}
	}
	fmt.Printf("Added user: %s\n", username)
}

func addServiceUser(conn *pgx.Conn, login string, password string, role string) {
	login = pgx.Identifier{login}.Sanitize()
	password = strings.ReplaceAll(password, "'", "''")
	query := fmt.Sprintf(`create user %s with password '%s' login;`, login, password)
	_, err := conn.Exec(context.Background(), query)
	if err != nil {
		pgErr, ok := errors.AsType[*pgconn.PgError](err)
		if ok {
			if pgErr.Code != "42710" {
				panic(err)
			}
		} else {
			panic(err)
		}
	}
	query = fmt.Sprintf(`grant %s to %s;`, role, login)
	_, err = conn.Exec(context.Background(), query)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Added service user %s\n", login)
}

func main() {
	envPaths := []string{
		".env",
		"../.env",
		"../../.env",
		"../../../.env",
	}
	var loaded bool
	for _, path := range envPaths {
		if _, err := os.Stat(path); err == nil {
			if err := godotenv.Load(path); err == nil {
				loaded = true
				break
			}
		}
	}
	if !loaded {
		panic("No .env files found")
	}

	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	name := os.Getenv("POSTGRES_DB")
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbUser, dbPassword, host, port, name)

	conn, err := pgx.Connect(context.Background(), dbUrl)
	if err != nil {
		panic(err)
	}
	defer func() {
		err = conn.Close(context.Background())
		if err != nil {
			fmt.Printf("Unable to close connection: %v\n", err)
		}
	}()

	defaultUserPassword := os.Getenv("DEFAULT_USER_PASSWORD")
	if defaultUserPassword == "" {
		panic("DEFAULT_USER_PASSWORD is not set")
	}
	addUser(conn, "test", defaultUserPassword, "test@example.com", false)
	defaultAdminPassword := os.Getenv("ADMIN_USER_PASSWORD")
	if defaultAdminPassword == "" {
		panic("ADMIN_USER_PASSWORD is not set")
	}
	addUser(conn, "admin", defaultAdminPassword, "admin@example.com", true)
	appServiceLogin := os.Getenv("APP_SERVICE_LOGIN")
	if appServiceLogin == "" {
		panic("APP_SERVICE_LOGIN is not set")
	}
	appServicePassword := os.Getenv("APP_SERVICE_PASSWORD")
	if appServicePassword == "" {
		panic("APP_SERVICE_PASSWORD is not set")
	}
	addServiceUser(conn, appServiceLogin, appServicePassword, "app_service_role")
	fileServiceLogin := os.Getenv("FILE_SERVICE_LOGIN")
	if fileServiceLogin == "" {
		panic("FILE_SERVICE_LOGIN is not set")
	}
	fileServicePassword := os.Getenv("FILE_SERVICE_PASSWORD")
	if fileServicePassword == "" {
		panic("FILE_SERVICE_PASSWORD is not set")
	}
	addServiceUser(conn, fileServiceLogin, fileServicePassword, "file_service_role")
	aiServiceLogin := os.Getenv("AI_SERVICE_LOGIN")
	if aiServiceLogin == "" {
		panic("AI_SERVICE_LOGIN is not set")
	}
	aiServicePassword := os.Getenv("AI_SERVICE_PASSWORD")
	if aiServicePassword == "" {
		panic("AI_SERVICE_PASSWORD is not set")
	}
	addServiceUser(conn, aiServiceLogin, aiServicePassword, "ai_service_role")
	authServiceLogin := os.Getenv("AUTH_SERVICE_LOGIN")
	if authServiceLogin == "" {
		panic("AUTH_SERVICE_LOGIN is not set")
	}
	authServicePassword := os.Getenv("AUTH_SERVICE_PASSWORD")
	if authServicePassword == "" {
		panic("AUTH_SERVICE_PASSWORD is not set")
	}
	addServiceUser(conn, authServiceLogin, authServicePassword, "auth_service_role")
}
