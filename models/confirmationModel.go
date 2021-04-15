package models

import (
	"os"
)

type Confirmation struct {
	Id             string `json:"id"`
	Activated      bool   `json:"activated"`
	Resend_expired int    `json:"resend_expired"`
	User_id        int    `json:"user_id"`
}

func SendEmailConfirm() {
	link := os.Getenv("APP_URL")
	_ = link

}
