package main

import (
	"fmt"
	"net/smtp"
	"os"
	"sync"
	"time"
)

// record macbook air m2, no html generation
// 10,000 emails, 1 goroutines, 57787 ms
// 10,000 emails, 5 goroutines, 38232 ms
// 10,000 emails, 10 goroutines, 37747 ms
// 10,000 emails, 20 goroutines, 37044 ms

// record macbook air m2, with html generation
// 10,000 emails, 1 goroutines, 245451 ms
// 10,000 emails, 5 goroutines, 50875 ms
// 10,000 emails, 10 goroutines, 35820 ms
// 10,000 emails, 20 goroutines, 34750 ms

func main() {
	// env
	smtpHost := getEnv("SMTP_HOST", "localhost")
	smtpPort := getEnv("SMTP_PORT", "25")
	// smtpUsername := getEnv("SMTP_USERNAME", "username")
	// smtpPassword := getEnv("SMTP_PASSWORD", "password")
	// auth := smtp.PlainAuth("identity", smtpUsername, smtpPassword, smtpHost)

	// goroutine amount
	goroutineLimit := 10
	limitChannel := make(chan int, goroutineLimit)
	var wg sync.WaitGroup

	// email amount
	emailAmount := 100
	if smtpHost != "localhost" {
		emailAmount = 1
	}
	fmt.Println("running with SMTP host", smtpHost, "with", emailAmount, "email(s), with", goroutineLimit, "concurrent goroutine(s)")

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
			headers := "X-NL-TYPE: nl_type_1\nX-CAMPAIGN: C1"
			// headers := "X-NL-TYPE: nl_type_1"
			message := []byte(
				fmt.Sprintf("Subject: %s\n%s\r\n\r\n%s", subject, headers, body),
			)
			err := smtp.SendMail(smtpHost+":"+smtpPort, nil, from, to, message)
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
			fakeGenerateHtmlTime()
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

func fakeGenerateHtmlTime() {
	// avg time when communicating through kube dns
	time.Sleep(15 * time.Millisecond)
}
