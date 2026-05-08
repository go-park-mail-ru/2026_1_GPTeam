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

func addBaseUser(conn *pgx.Conn) {
	username := "test"
	plainPassword := os.Getenv("DEFAULT_USER_PASSWORD")
	bytes, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("Unable to hash DEFAULT_USER_PASSWORD: %v\n", err)
		return
	}
	password := string(bytes)
	email := "test@example.com"
	query := `insert into "user" (username, password, email) VALUES ($1, $2, $3);`
	_, err = conn.Exec(context.Background(), query, username, password, email)
	if err != nil {
		pgErr, ok := errors.AsType[*pgconn.PgError](err)
		if ok {
			if pgErr.Code != "23505" {
				fmt.Printf("Unable to execute sql: %v\n", err)
				panic(err)
			}
		} else {
			fmt.Printf("Unable to execute sql: %v\n", err)
			panic(err)
		}
	}
	fmt.Println("Added base user")
}

func addAdminUser(conn *pgx.Conn) {
	username := "admin"
	plainPassword := os.Getenv("ADMIN_USER_PASSWORD")
	bytes, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("Unable to hash ADMIN_USER_PASSWORD: %v\n", err)
		return
	}
	password := string(bytes)
	email := "admin@example.com"
	query := `insert into "user" (username, password, email, is_staff) VALUES ($1, $2, $3, true);`
	_, err = conn.Exec(context.Background(), query, username, password, email)
	if err != nil {
		pgErr, ok := errors.AsType[*pgconn.PgError](err)
		if ok {
			if pgErr.Code != "23505" {
				fmt.Printf("Unable to execute sql: %v\n", err)
				panic(err)
			}
		} else {
			fmt.Printf("Unable to execute sql: %v\n", err)
			panic(err)
		}
	}
	fmt.Println("Added admin user")
}

func addAppServiceUser(conn *pgx.Conn) {
	login := os.Getenv("APP_SERVICE_LOGIN")
	password := os.Getenv("APP_SERVICE_PASSWORD")
	login = strings.ReplaceAll(login, "'", "''")
	password = strings.ReplaceAll(password, "'", "''")
	query := fmt.Sprintf(`create user %s with password '%s' login;`, login, password)
	_, err := conn.Exec(context.Background(), query)
	if err != nil {
		pgErr, ok := errors.AsType[*pgconn.PgError](err)
		if ok {
			if pgErr.Code != "42710" {
				fmt.Printf("Unable to execute sql: %v\n", err)
				panic(err)
			}
		} else {
			fmt.Printf("Unable to execute sql: %v\n", err)
			panic(err)
		}
	}
	query = fmt.Sprintf(`grant app_service_role to %s;`, login)
	_, err = conn.Exec(context.Background(), query)
	if err != nil {
		pgErr, ok := errors.AsType[*pgconn.PgError](err)
		if ok {
			if pgErr.Code != "42710" {
				fmt.Printf("Unable to execute sql: %v\n", err)
				panic(err)
			}
		} else {
			fmt.Printf("Unable to execute sql: %v\n", err)
			panic(err)
		}
	}
	fmt.Println("Added app service user")
}

func addFileServiceUser(conn *pgx.Conn) {
	login := os.Getenv("FILE_SERVICE_LOGIN")
	password := os.Getenv("FILE_SERVICE_PASSWORD")
	login = strings.ReplaceAll(login, "'", "''")
	password = strings.ReplaceAll(password, "'", "''")
	query := fmt.Sprintf(`create user %s with password '%s' login;`, login, password)
	_, err := conn.Exec(context.Background(), query)
	if err != nil {
		pgErr, ok := errors.AsType[*pgconn.PgError](err)
		if ok {
			if pgErr.Code != "42710" {
				fmt.Printf("Unable to execute sql: %v\n", err)
				panic(err)
			}
		} else {
			fmt.Printf("Unable to execute sql: %v\n", err)
			panic(err)
		}
	}
	query = fmt.Sprintf(`grant file_service_role to %s;`, login)
	_, err = conn.Exec(context.Background(), query)
	if err != nil {
		pgErr, ok := errors.AsType[*pgconn.PgError](err)
		if ok {
			if pgErr.Code != "42710" {
				fmt.Printf("Unable to execute sql: %v\n", err)
				panic(err)
			}
		} else {
			fmt.Printf("Unable to execute sql: %v\n", err)
			panic(err)
		}
	}
	fmt.Println("Added file service user")
}

func addAiServiceUser(conn *pgx.Conn) {
	login := os.Getenv("AI_SERVICE_LOGIN")
	password := os.Getenv("AI_SERVICE_PASSWORD")
	login = strings.ReplaceAll(login, "'", "''")
	password = strings.ReplaceAll(password, "'", "''")
	query := fmt.Sprintf(`create user %s with password '%s' login;`, login, password)
	_, err := conn.Exec(context.Background(), query)
	if err != nil {
		pgErr, ok := errors.AsType[*pgconn.PgError](err)
		if ok {
			if pgErr.Code != "42710" {
				fmt.Printf("Unable to execute sql: %v\n", err)
				panic(err)
			}
		} else {
			fmt.Printf("Unable to execute sql: %v\n", err)
			panic(err)
		}
	}
	query = fmt.Sprintf(`grant ai_service_role to %s;`, login)
	_, err = conn.Exec(context.Background(), query)
	if err != nil {
		pgErr, ok := errors.AsType[*pgconn.PgError](err)
		if ok {
			if pgErr.Code != "42710" {
				fmt.Printf("Unable to execute sql: %v\n", err)
				panic(err)
			}
		} else {
			fmt.Printf("Unable to execute sql: %v\n", err)
			panic(err)
		}
	}
	fmt.Println("Added ai service user")
}

func addAuthServiceUser(conn *pgx.Conn) {
	login := os.Getenv("AUTH_SERVICE_LOGIN")
	password := os.Getenv("AUTH_SERVICE_PASSWORD")
	login = strings.ReplaceAll(login, "'", "''")
	password = strings.ReplaceAll(password, "'", "''")
	query := fmt.Sprintf(`create user %s with password '%s' login;`, login, password)
	_, err := conn.Exec(context.Background(), query)
	if err != nil {
		pgErr, ok := errors.AsType[*pgconn.PgError](err)
		if ok {
			if pgErr.Code != "42710" {
				fmt.Printf("Unable to execute sql: %v\n", err)
				panic(err)
			}
		} else {
			fmt.Printf("Unable to execute sql: %v\n", err)
			panic(err)
		}
	}
	query = fmt.Sprintf(`grant auth_service_role to %s;`, login)
	_, err = conn.Exec(context.Background(), query)
	if err != nil {
		pgErr, ok := errors.AsType[*pgconn.PgError](err)
		if ok {
			if pgErr.Code != "42710" {
				fmt.Printf("Unable to execute sql: %v\n", err)
				panic(err)
			}
		} else {
			fmt.Printf("Unable to execute sql: %v\n", err)
			panic(err)
		}
	}
	fmt.Println("Added auth service user")
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
		fmt.Println("No .env files found")
		return
	}

	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	name := os.Getenv("POSTGRES_DB")
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbUser, dbPassword, host, port, name)

	conn, err := pgx.Connect(context.Background(), dbUrl)
	if err != nil {
		fmt.Printf("Unable to connect to database: %v\n", err)
		return
	}
	defer func() {
		err = conn.Close(context.Background())
		if err != nil {
			fmt.Printf("Unable to close connection: %v\n", err)
		}
	}()

	addBaseUser(conn)
	addAdminUser(conn)
	addAppServiceUser(conn)
	addAuthServiceUser(conn)
	addAiServiceUser(conn)
	addFileServiceUser(conn)
}
