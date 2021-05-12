package config

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
	"os"
)

func InitDB() *sql.DB {
	dbURI := os.Getenv("POSTGRES_URL")

	db, err := sql.Open("postgres", dbURI)

	if err != nil {
		log.Fatal("Failed connect to database: ", err)
	}
	return db
}
