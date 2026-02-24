package services

import (
	"log"
	"net/smtp"
)

func SendEmail(to, subject, body string) {
	from := "ustjapanesemyanamarengineering@gmail.com"
	password := "sstvrrhrbzqcgznc"

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	err := smtp.SendMail(
		"smtp.gmail.com:587",
		smtp.PlainAuth("", from, password, "smtp.gmail.com"),
		from,
		[]string{to},
		[]byte(msg),
	)
	if err != nil {
		log.Println("[ERROR] SendEmail failed:", err)
	} else {
		log.Println("[DEBUG] Email sent to:", to)
	}
}
