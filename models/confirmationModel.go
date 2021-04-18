package models

import (
	// "github.com/google/uuid"
	"golangblog/database"
	"golangblog/libs"
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

	stmt, _ := db.Prepare(query)

	stmt.QueryRow(user_id)
	return nil
}

func FindConfirmation(user_id uint64) (Confirmation, error) {
	var confirmation Confirmation

	db := database.InitDB()
	defer db.Close()

	query := `select activated from confirmation_users where user_id = $1`

	row := db.QueryRow(query, user_id)
	row.Scan(&confirmation.Activated)

	return confirmation, nil
}
