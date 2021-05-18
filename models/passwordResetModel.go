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

func CheckEmailPasswordReset(email string) int {
	var counter int

	db := config.InitDB()
	defer db.Close()

	query := `select count(id) from password_resets where email = $1`

	db.QueryRow(query, email).Scan(&counter)

	return counter
}

func CheckResendExpired(email string) int {
	var counter int

	db := config.InitDB()
	defer db.Close()

	query := `select count(resend_expired) from password_resets where email = $1`

	db.QueryRow(query, email).Scan(&counter)

	return counter
}

func GetResendExpired(email string) PasswordReset {
	var passwordreset PasswordReset

	db := config.InitDB()
	defer db.Close()

	query := `select resend_expired from password_resets where email = $1`

	row := db.QueryRow(query, email)
	row.Scan(&passwordreset.ResendExpired)

	return passwordreset
}

func ChangeResendExpired(email string) {
	db := config.InitDB()
	defer db.Close()

	query := `update password_resets set resend_expired = $1 where email = $2`

	stmt, _ := db.Prepare(query)

	resend := GetResendExpired(email)
	changeResend := resend.ResendExpired + 300
	stmt.Exec(changeResend, email)

	return
}
