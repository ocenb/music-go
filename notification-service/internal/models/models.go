package models

type EmailNotification struct {
	Email string `json:"email"`
	Msg   string `json:"msg"`
}
