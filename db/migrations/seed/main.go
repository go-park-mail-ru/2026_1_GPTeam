package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

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

	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	name := os.Getenv("POSTGRES_DB")
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, name)

	conn, err := pgx.Connect(context.Background(), dbUrl)
	if err != nil {
		fmt.Printf("Unable to connect to database: %v\n", err)
		return
	}

	adminUsername := pgtype.Text{
		String: "admin",
		Valid:  true,
	}
	plainPassword := os.Getenv("DEFAULT_USER_PASSWORD")
	bytes, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("Unable to hash password: %v\n", err)
		return
	}
	adminPassword := pgtype.Text{
		String: string(bytes),
		Valid:  true,
	}
	adminEmail := pgtype.Text{
		String: "admin@example.com",
		Valid:  true,
	}
	adminLastLogin := pgtype.Timestamp{
		Time:  time.Time{},
		Valid: false,
	}
	adminAvatar := pgtype.Text{
		String: "img/123.png",
		Valid:  true,
	}
	addUserSQL := "insert into \"user\" (username, password, email, last_login, avatar_url) VALUES ($1, $2, $3, $4, $5);"

	_, err = conn.Exec(context.Background(), addUserSQL, adminUsername, adminPassword, adminEmail, adminLastLogin, adminAvatar)
	if err != nil {
		fmt.Printf("Unable to execute sql: %v\n", err)
		return
	}
	fmt.Println("Added admin user")

	defer func() {
		err = conn.Close(context.Background())
		if err != nil {
			fmt.Printf("Unable to close connection: %v\n", err)
		}
	}()
}
