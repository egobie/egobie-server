package controllers

import (
	"fmt"

	"github.com/njern/gonexmo"

	"github.com/egobie/egobie-server/config"
)

func SendCancelMessage(to string) {
	sendSms(to, "Due to weather problem, we may cancel your service")
}

func sendSms(to, message string) {
	smsMessage := &nexmo.SMSMessage{
		From: config.SmsFrom,
		To:   "1" + to,
		Type: nexmo.Text,
		Text: message,
	}

	if _, err := config.Sms.SMS.Send(smsMessage); err != nil {
		fmt.Println("Error when sending message - ", err.Error())
	}
}
