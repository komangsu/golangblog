package models

import (
	// "github.com/google/uuid"
	"golangblog/database"
	"golangblog/libs"
	"log"
	"os"
	"runtime"
)

type Confirmation struct {
	Id             uint64 `json:"id"`
	Activated      bool   `json:"activated"`
	Resend_expired uint64 `json:"resend_expired"`
	User_id        uint64 `json:"user_id"`
}

type bodylink struct {
	Name string
	URL  string
}

func SendEmailConfirm(username, email string) {
	link := os.Getenv("APP_URL") + "/test"

	templateData := bodylink{
		Name: username,
		URL:  link,
	}

	runtime.GOMAXPROCS(1)
	go libs.SendEmailVerification(email, templateData)
}

func SaveConfirmation(user_id uint64) error {

	db := database.InitDB()
	defer db.Close()

	query := `insert into confirmation_users(user_id) values($1)`

	stmt, err := db.Prepare(query)
	if err != nil {
		panic(err)
	}

	queryErr := stmt.QueryRow(user_id)
	if queryErr != nil {
		log.Fatal(queryErr)
	}

	return nil
}
