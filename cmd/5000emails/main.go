package main

import (
	"fmt"
	"net/smtp"
	"sync"
	"time"
)

func main() {
	// env
	smtpHost := "gosmtp.nv-development.svc.cluster.local"
	smtpPort := "25"

	// goroutine amount
	goroutineLimit := 10
	limitChannel := make(chan int, goroutineLimit)
	var wg sync.WaitGroup

	// email amount
	emailAmount := 5000
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
			from := "from@gmail.com"
			to := []string{"to@gmail.com"}
			subject := "Test subject"
			body := fmt.Sprintf("Test body: %d", i)
			message := []byte(
				fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, body),
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

func fakeGenerateHtmlTime() {
	// avg time when communicating through kube dns
	time.Sleep(15 * time.Millisecond)
}
