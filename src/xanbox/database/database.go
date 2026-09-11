package database

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/uptrace/bun"
)

func NewPool() *bun.DB {
	godotenv.Load()

	dsn := os.Getenv("DATABASE_URL")

	return ConnectDB(dsn)
}
