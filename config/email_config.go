package config

import (
	"fmt"
	"net/smtp"
	"os"
)

var (
	EmailUsername string
	EmailPassword string
	EmailHost     string
	EmailPort     string
	EmailAddress  string
	EmailSender   string
)

var Email smtp.Auth

func init() {
	os.Getenv("")

	EmailUsername = os.Getenv("EGOBIE_EMAIL_USERNAME")
	EmailPassword = os.Getenv("EGOBIE_EMAIL_PASSWORD")
	EmailHost = os.Getenv("EGOBIE_EMAIL_HOST")
	EmailPort = os.Getenv("EGOBIE_EMAIL_PORT")
	EmailSender = os.Getenv("EGOBIE_EMAIL_SENDER")

	if EmailUsername == "" || EmailPassword == "" || EmailPort == "" ||
		EmailSender == "" || EmailHost == "" {
		fmt.Println("Email not configured properly")
		os.Exit(1)
	}

	Email = smtp.PlainAuth(
		"",
		EmailUsername,
		EmailPassword,
		EmailHost,
	)

	EmailAddress = EmailHost + ":" + EmailPort
}
