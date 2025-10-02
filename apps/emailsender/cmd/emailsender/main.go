package main

// NOTE: this verson is really slow because smtp clients are not reused

import (
	"fmt"
	"net/smtp"
	"sync"
	"time"

	"github.com/kiart-tantasi/email-sender-go/internal/env"
	"github.com/kiart-tantasi/email-sender-go/internal/fake"
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
	smtpHost := env.GetEnv("SMTP_HOST", "localhost")
	smtpPort := env.GetEnv("SMTP_PORT", "25")
	// smtpUsername := env.GetEnv("SMTP_USERNAME", "username")
	// smtpPassword := env.GetEnv("SMTP_PASSWORD", "password")
	// auth := smtp.PlainAuth("identity", smtpUsername, smtpPassword, smtpHost)

	// goroutine amount
	goroutineLimit := 10
	limitChannel := make(chan int, goroutineLimit)
	var wg sync.WaitGroup

	// email amount
	emailAmount := 10000
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
			fake.FakeGenerateHtmlTime()
			<-limitChannel
		}(i)
	}
	wg.Wait()

	fmt.Println("done in", time.Since(start).Milliseconds(), "ms")
	fmt.Println("errCount:", errCount)
	fmt.Println("sucessCount:", successCount)
}
