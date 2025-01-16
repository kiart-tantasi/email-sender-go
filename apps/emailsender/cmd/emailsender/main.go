package main

import (
	"fmt"
	"net/smtp"
	"os"
)

func main() {
	smtpHost := "localhost"
	smtpPort := "25"

	from := "testfrom@example.com"
	to := []string{"testto@example.com"}
	subject := "Test subject"
	body := "Test body."
	message := []byte(
		fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, body),
	)

	auth := smtp.PlainAuth("identity", "username", "password", smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message)
	if err != nil {
		fmt.Println("Failed to send email:", err)
		os.Exit(1)
	}

	fmt.Println("Email sent successfully!")
}
