package repository

//go:generate mockgen -destination=mocks/pgx.go -package=mocks github.com/jackc/pgx/v5 Row,Rows
