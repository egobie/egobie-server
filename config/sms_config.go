package config

import (
	"fmt"
	"os"

	"github.com/njern/gonexmo"
)

var (
	SmsApiKey    string
	SmsApiSecret string
	SmsFrom      string
)

var Sms *nexmo.Client

func init() {
	var err error

	os.Getenv("")

	SmsApiKey = os.Getenv("EGOBIE_SMS_API_KEY")
	SmsApiSecret = os.Getenv("EGOBIE_SMS_API_SECRET")
	SmsFrom = os.Getenv("EGOBIE_SMS_FROM")

	if SmsApiKey == "" || SmsApiSecret == "" {
		fmt.Println("SMS not configured properly")
		os.Exit(0)
	}

	if Sms, err = nexmo.NewClientFromAPI(
		SmsApiKey, SmsApiSecret,
	); err != nil {
		fmt.Println("SMS not initialized properly : ", err.Error())
	}
}
