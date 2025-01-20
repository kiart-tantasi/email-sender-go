package main

import (
	"fmt"
	"net/smtp"
	"os"
	"sync"
	"time"
)

// record macbook air m2
// 10,000 emails, 1 goroutines, 57787 ms
// 10,000 emails, 5 goroutines, 38232 ms
// 10,000 emails, 10 goroutines, 37747 ms
// 10,000 emails, 20 goroutines, 37044 ms

// 100,000 emails, 5 goroutines, 379796 ms
// 100,000 emails, 10 goroutines, 365820 ms

func main() {
	smtpHost := getEnv("SMTP_HOST", "localhost")
	smtpPort := getEnv("SMTP_PORT", "25")
	smtpUsername := getEnv("SMTP_USERNAME", "username")
	smtpPassword := getEnv("SMTP_PASSWORD", "password")
	auth := smtp.PlainAuth("identity", smtpUsername, smtpPassword, smtpHost)

	// goroutine config
	goroutineLimit := 5
	limitChannel := make(chan int, goroutineLimit)
	var wg sync.WaitGroup

	// email amount
	emailAmount := 1_000
	if smtpHost != "localhost" {
		emailAmount = 1
	}
	fmt.Println("running with SMTP host", smtpHost, "with", emailAmount, "email(s)")

	successCount := 0
	errCount := 0
	start := time.Now()
	for i := 0; i < emailAmount; i++ {
		wg.Add(1)
		limitChannel <- 1
		go func(i int) {
			defer wg.Done()

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
			} else {
				successCount++
			}

			// log progress
			if i%10 == 0 {
				fmt.Println("Sent email index", i)
			}
			<-limitChannel
		}(i)
	}
	wg.Wait()

	fmt.Println("done in", time.Since(start).Milliseconds(), "ms")
	fmt.Println("errCount:", errCount)
	fmt.Println("sucessCount:", successCount)
}

func getEnv(envName, defaultValue string) string {
	val := os.Getenv(envName)
	if val != "" {
		return val
	}
	return defaultValue
}
