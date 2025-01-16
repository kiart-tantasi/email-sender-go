package main

import (
	"fmt"
	"net/smtp"
	"os"
	"time"
)

func main() {
	smtpHost := getEnv("SMTP_HOST", "localhost")
	smtpPort := getEnv("SMTP_PORT", "25")
	smtpUsername := getEnv("SMTP_USERNAME", "username")
	smtpPassword := getEnv("SMTP_PASSWORD", "password")
	auth := smtp.PlainAuth("identity", smtpUsername, smtpPassword, smtpHost)

	fmt.Println("running with SMTP host", smtpHost)
	errCount := 0
	start := time.Now()
	for i := 0; i < 100; i++ {
		// send email
		from := "kiarttantasi@gmail.com"
		to := []string{"kiarttantasi@gmail.com"}
		subject := "Test subject"
		body := fmt.Sprintf("Test body: %d", i)
		message := []byte(
			fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, body),
		)
		err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message)
		if err != nil {
			errCount++
			fmt.Println(err)
		}

		// progress
		if i%10 == 0 {
			fmt.Println("Sent email index", i)
		}
	}

	fmt.Println("done in", time.Since(start).Milliseconds())
	fmt.Println("errCount:", errCount)
}

func getEnv(envName, defaultValue string) string {
	val := os.Getenv(envName)
	if val != "" {
		return val
	}
	return defaultValue
}
