package controllers

import (
	"fmt"
	"bytes"
	"net/smtp"
	"text/template"

	"github.com/egobie/egobie-server/config"
	"github.com/egobie/egobie-server/modules"
)

func sendPlaceOrderEmail(
	address, name, reservationNumber, startTime string,
	services []string, addons []string,
	cost float32) {

	message := "Hello " + name + ",\n" +
		"\n" +
		"This letter is confirmation of your reservation at eGobie Car Services. " +
		"Please see the details of the reservation noted below:\n" +
		"\n" +
		"Reservation Number: " + reservationNumber + "\n" +
		"Estimated Start: " + startTime + "\n" +
		"Total Cost: $" + fmt.Sprintf("%v", cost) + "\n"

	if len(services) > 0 {
		message += "Services: \n"

		for _, service := range services {
			message += " - " + service + "\n"
		}
	}

	if len(addons) > 0 {
		message += "Extra Services: \n"

		for _, addon := range addons {
			message += " - " + addon + "\n"
		}
	}

	message += "\n" +
		"We only process the payment after the service is done. We require you to cancel " +
		"the service appointment 24 hours ahead, otherwise we will charge 50% of the appointment cost. " +
		"If we show up at the door and no one is around, we will charge 100% of the appointment cost " +
		"as no-show fee.\n" +
		"\n" +
		"Thank you for using eGobie Car Services\n"

	email := &modules.EmailTemplate{
		config.EmailSender,
		address,
		"Thanks for using eGobie",
		message,
	}
	content := "From: eGobie Car Services <{{.From}}>\n" +
		"To: {{.To}}\n" +
		"Subject: {{.Subject}}\n" +
		"\n" +
		"{{.Body}}"

	var (
		t *template.Template
		err error
		doc bytes.Buffer
	)

	if t, err = template.New("template").Parse(content); err != nil {
		fmt.Println("Error - Parse - ", err.Error())
	}

	if err = t.Execute(&doc, email); err != nil {
		fmt.Println("Error - Execute - ", err.Error())
	}

	if err = smtp.SendMail(
		config.EmailAddress,
		config.Email,
		config.EmailSender,
		[]string{address},
		doc.Bytes(),
	); err != nil {
		fmt.Println("Error - Email - PlaceOrder: ", err.Error())
	}
}

func sendNewFleetUserEmail(address, name, token string) {

	message := "Hello " + name + ",\n" +
		"\n" +
		"Thank you for using eGobie Car Services. " +
		"Please use following email address and token to sign up eGobie fleet app:\n" +
		"\n" +
		"Email Address: " + address + "\n" +
		"Sign-Up Token: " + token + "\n" +
		"\n" +
		"Thank you for using eGobie car services\n"

	email := &modules.EmailTemplate{
		config.EmailSender,
		address,
		"[New Fleet Service User] Thanks for using eGobie",
		message,
	}
	content := "From: eGobie Car Services <{{.From}}>\n" +
		"To: {{.To}}\n" +
		"Subject: {{.Subject}}\n" +
		"\n" +
		"{{.Body}}"

	var (
		t *template.Template
		err error
		doc bytes.Buffer
	)

	if t, err = template.New("template").Parse(content); err != nil {
		fmt.Println("Error - Parse - ", err.Error())
	}

	if err = t.Execute(&doc, email); err != nil {
		fmt.Println("Error - Execute - ", err.Error())
	}

	if err = smtp.SendMail(
		config.EmailAddress,
		config.Email,
		config.EmailSender,
		[]string{address, config.EmailCEO},
		doc.Bytes(),
	); err != nil {
		fmt.Println("Error - Email - NewFleetUser: ", err.Error())
	}
}
