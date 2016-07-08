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
	services []string, addons []string, cost float32) {

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

	sendEmail(
		address, "[New Residential Reservation] Thanks for using eGobie",
		message, false, true,
	)
}

func sendNewResidentialUserEmail(address string) {
	message := `
		<html><body>
		<p style="font-size: 17px">Welcome to eGobie!  We hope you will enjoy our services and help us spread the word if you are satisfied. </p>
		<ul style="font-size: 13px; margin-bottom: 20px">
			<li style="margin-bottom: 5px">As a promotion, we will take <u><b>50% off your first service order</b></u>, automatically taken out during reservation.  We want you to try out our service with a peace of mind.</li>
			<li style="margin-bottom: 5px">If you order both oil change and car wash service, you will receive another 10% discount!</li>
			<li style="margin-bottom: 5px">Combining the 2 offers, you can receive up to 55% (cumulative) discount off your first order!</li>
		</ul>
		<p style="font-size: 17px">Want an Extra 10% Off Starting on the Second Order?</p>
		<p style="font-size: 13px; margin-bottom: 30px">Check the <u><b>Get Your Gift</b></u> Section on your App - Share "Your Code" with your friends via text or Social Media and Get another <u><b>10% Off</b></u> when they use your code to register an eGobie account! (Limit 1 certificate per order.  10% discount will be automatically deducted starting from your second order)</p>
		<p style="font-size: 13px">Welcome to eGobie, we bring you time, quality, and convenience.</p>
		<p style="font-size: 13px">eGobie Team<p>
		</body></html>
	`

	sendEmail(
		address, "Up to 55% off for First Time eGobie Users!",
		message, true, true,
	)
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

	sendEmail(
		address, "[New Fleet Service User] Thanks for using eGobie",
		message, false, true,
	)
}

func sendResetPasswordEmail(address, name, token string) {
	message := "Hello " + name + ",\n" +
		"\n" +
		"Thank you for using eGobie Car Services. " +
		"Please use following token to reset your eGobie password:\n" +
		"\n" +
		"Token: " + token + "\n" +
		"\n" +
		"Thank you for using eGobie car services\n"

	sendEmail(
		address, "Reset your eGobie password", message, false, false,
	)
}

func sendEmail(address, subject, body string, html, sendToCEO bool) {
	email := &modules.EmailTemplate{
		config.EmailSender,
		address,
		subject,
		body,
	}
	addrs := []string {address}

	if sendToCEO {
		addrs = append(addrs, config.EmailCEO)
	}

	content := "From: eGobie Car Services <{{.From}}>\n" +
		"To: {{.To}}\n" +
		"Subject: {{.Subject}}\n"

	if html {
		content += "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n"
	}

	content += "\n{{.Body}}"

	var (
		t *template.Template
		err error
		doc bytes.Buffer
	)

	if t, err = template.New("template").Parse(content); err != nil {
		fmt.Println("Error - Parse Template - ", err.Error())
	}

	if err = t.Execute(&doc, email); err != nil {
		fmt.Println("Error - Execute Template - ", err.Error())
	}

	if err = smtp.SendMail(
		config.EmailAddress,
		config.Email,
		config.EmailSender,
		addrs,
		doc.Bytes(),
	); err != nil {
		fmt.Println("Error - Send Email - ", err.Error())
	}
}
