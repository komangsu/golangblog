package models

import (
	"golangblog/config"
	"time"
)

type PasswordReset struct {
	Id            uint64    `json:"id"`
	Email         string    `json:"email"`
	ResendExpired uint64    `json:"resend_expired"`
	CreatedAt     time.Time `json:"created_at"`
}

func SavePasswordReset(email string) {
	db := config.InitDB()
	defer db.Close()

	query := `insert into password_resets(email,resend_expired) values($1,$2)`

	stmt, _ := db.Prepare(query)

	resend_expired := time.Now().Add(time.Minute * 5).Unix() // set expired 5 minutes
	stmt.QueryRow(email, resend_expired)

	return
}
