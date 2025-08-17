package main

import (
	"fmt"
	"log"
	"net/smtp"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kiart-tantasi/email-sender-go/internal/env"
)

/*
[RESULTS]

100 sent, 1 goroutines, 100 channel capacity, 476 ms
100 sent, 10 goroutines, 100 channel capacity, 343 ms
100 sent, 100 goroutines, 100 channel capacity, 271 ms
1000 sent, 1 goroutines, 100 channel capacity, 4760 ms
1000 sent, 10 goroutines, 100 channel capacity, 2981 ms
1000 sent, 100 goroutines, 100 channel capacity, 2919 ms
10000 sent, 1 goroutines, 100 channel capacity, 66930 ms
10000 sent, 10 goroutines, 100 channel capacity, 46406 ms
10000 sent, 100 goroutines, 100 channel capacity, 44224 ms
*/
type Queue struct {
	from string
	to   []string
	msg  []byte
}

// CONFIG
var EMAIL_COUNT int = 10_000
var GOROUTINE_COUNT = 1
var CHANNEL_CAPACITY int = 100

func main() {
	smtpHost := env.GetEnv("SMTP_HOST", "localhost")
	smtpPort := env.GetEnv("SMTP_PORT", "25")

	queue := make(chan Queue, CHANNEL_CAPACITY)
	var wg sync.WaitGroup
	var sentCount int32 = 0
	start := time.Now()

	// enqueue mocks
	wg.Add(1)
	go func() {
		from, to, msg := mockEmailPaylod()
		for range EMAIL_COUNT {
			queue <- Queue{
				from,
				to,
				msg,
			}
			wg.Add(1)
		}
		wg.Done()
	}()

	// create multiple goroutines to send email to smtp server
	for i := range GOROUTINE_COUNT {
		go func() {
			log.Printf("Started goroutine %d", i)

			// infinite loop
			for {
				select {
				case q := <-queue:
					if err := smtp.SendMail(smtpHost+":"+smtpPort, nil, q.from, q.to, q.msg); err != nil {
						log.Printf("Error when sending email on goroutine %d: %s", i, err)
					}
					atomic.AddInt32(&sentCount, 1)
					wg.Done()
				case <-time.After(10 * time.Second):
					log.Println("Error: No message received for 10 seconds or more.")
				}
			}

		}()
	}

	// wait
	wg.Wait()
	fmt.Printf("%d sent, %d goroutines, %d channel capacity, %d ms\n", sentCount, GOROUTINE_COUNT, CHANNEL_CAPACITY, time.Since(start).Milliseconds())
}

func mockEmailPaylod() (string, []string, []byte) {
	from := "from@test.com"
	to := []string{"to@test.com"}
	msg := "Subject: Hello from Example.com !\n\nThis is body"
	return from, to, []byte(msg)
}
