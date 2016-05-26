package controllers

import (
	"fmt"
	"net/smtp"
	"github.com/egobie/egobie-server/config"
)

func sendPlaceOrderEmail(address []string, message []byte) {
	if err := smtp.SendMail(
		config.EmailAddress,
		config.Email,
		config.EmailSender,
		address,
		message,
	); err != nil {
		fmt.Println("Error - Email - PlaceOrder: ", err.Error())
	} else {
		fmt.Println("Send Email to ", address)
	}
}
