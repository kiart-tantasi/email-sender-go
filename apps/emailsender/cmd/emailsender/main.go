package main

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kiart-tantasi/email-sender-go/internal/env"
)

// NOTE: there is smtp client pool but smtp clien is still always created. if you have time, please fix it.

/*
[RESULTS]

(macbook air m2)

# 100

100 send, 1 goroutines, - smtp  clients, 100 channel capacity, 476 ms, (successful send: 100)
100 send, 1 goroutines, 1 smtp  clients, 100 channel capacity, 297 ms, (successful send: 100)
100 send, 10 goroutines, - smtp  clients, 100 channel capacity, 343 ms, (successful send: 100)
100 send, 10 goroutines, 10 smtp  clients, 100 channel capacity, 180 ms, (successful send: 100)
100 send, 100 goroutines, - smtp  clients, 100 channel capacity, 271 ms, (successful send: 100)
100 send, 100 goroutines, 100 smtp  clients, 100 channel capacity, 243 ms, (successful send: 100)

# 1,000

1000 send, 1 goroutines, - smtp clients, 100 channel capacity, 4760 ms, (successful send: 1000)
1000 send, 1 goroutines, 1 smtp clients, 100 channel capacity, 2255 , (successful send: 1000)
1000 send, 10 goroutines, - smtp clients, 100 channel capacity, 2981 , (successful send: 1000)
1000 send, 10 goroutines, 10 smtp clients, 100 channel capacity, 1571 , (successful send: 1000)
1000 send, 100 goroutines, - smtp clients, 100 channel capacity, 2919 , (successful send: 1000)
1000 send, 100 goroutines, 100 smtp clients, 100 channel capacity, 1601 , (successful send: 1000)

# 10,000

10000 send, 1 goroutines, - smtp clients, 100 channel capacity, 66930 ms, (successful send: 10000)
10000 send, 1 goroutines, 1 smtp clients, 100 channel capacity, 20617 ms, (successful send: 10000)
10000 send, 10 goroutines, - smtp clients, 100 channel capacity, 46406 ms, (successful send: 10000)
10000 send, 10 goroutines, 10 smtp clients, 100 channel capacity, 15192 ms, (successful send: 10000)
10000 send, 100 goroutines, - smtp clients, 100 channel capacity, 44224 ms, (successful send: 10000)
10000 send, 100 goroutines, 100 smtp clients, 100 channel capacity, 17212 ms, (successful send: 10000)

# 100,000

100000 send, 1 goroutines, 1 smtp clients, 100 channel capacity, 195919 ms, (successful send: 100000)
100000 send, 10 goroutines, 10 smtp  clients, 100 channel capacity, 157515 ms, (successful send: 100000)
100000 send, 100 goroutines, 100 smtp clients, 100 channel capacity, 161958 ms, (successful send: 100000)
*/
type Queue struct {
	from string
	to   []string
	msg  []byte
}

// CONFIG
var EMAIL_COUNT int = 100_000
var GOROUTINE_COUNT = 10
var CHANNEL_CAPACITY int = 100

// CONUTERS
var smtpClientCounter int32 = 0
var sendCounter int32 = 0
var successfulSendCounter int32 = 0

func main() {
	smtpHost := env.GetEnv("SMTP_HOST", "localhost")
	smtpPort := env.GetEnv("SMTP_PORT", "25")

	queue := make(chan Queue, CHANNEL_CAPACITY)
	var wg sync.WaitGroup
	start := time.Now()

	// enqueue mocks
	wg.Add(1)
	go func() {
		defer wg.Done()
		from, to, msg := mockEmailPaylod()
		for range EMAIL_COUNT {
			queue <- Queue{
				from,
				to,
				msg,
			}
			wg.Add(1)
		}
	}()

	smtpClientPool := &SmtpClientPool{}

	// create multiple goroutines to send email to smtp server
	for i := range GOROUTINE_COUNT {
		go func(idx int) {
			log.Printf("Started goroutine %d", idx)
			client, err := smtpClientPool.acquire(smtpHost, smtpPort)
			if err != nil {
				log.Printf("Errror when acquiring smtp client: %s", err)
				return
			}

			// infinite loop
			for {
				select {
				case q := <-queue:

					client.Hello("TEST")
					client.Mail(q.from)
					client.Rcpt(q.to[0])
					writer, err := client.Data()
					// reconnect if client gets error
					if err != nil {
						log.Printf("Client data error, attempting to reconnect: %v", err)
						client, err = newClient(smtpHost, smtpPort) // temp way to reconnect
						if err != nil {
							log.Printf("Error when reconnecting smtp client: %s", err)
							return
						}
					}

					// Send email
					if _, err := writer.Write(q.msg); err != nil {
						log.Printf("Error when sending email on goroutine %d: %s", i, err)
						continue
					} else {
						writer.Close()
						atomic.AddInt32(&successfulSendCounter, 1)
					}

					atomic.AddInt32(&sendCounter, 1)
					wg.Done()
				case <-time.After(10 * time.Second):
					log.Println("Error: No message received for 10 seconds so app wil exit")
					os.Exit(1)
				}
			}

		}(i)
	}

	// wait
	wg.Wait()
	fmt.Printf("%d send, %d goroutines, %d smtp clients, %d channel capacity, %d ms, (successful send: %d)\n", sendCounter, GOROUTINE_COUNT, smtpClientCounter, CHANNEL_CAPACITY, time.Since(start).Milliseconds(), successfulSendCounter)
}

func mockEmailPaylod() (string, []string, []byte) {
	from := "from@test.com"
	to := []string{"to@test.com"}
	msg := "Subject: Hello from Example.com !\n\nThis is body"
	return from, to, []byte(msg)
}

func newClient(smtpHost, smtpPort string) (*smtp.Client, error) {
	atomic.AddInt32(&smtpClientCounter, 1)
	addr := smtpHost + ":" + smtpPort
	client, err := smtp.Dial(addr)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// smtp client pool
// TODO: move into a separate package
type SmtpClientPoolI interface {
	acquire(smtpHost, smtpPort string) (*smtp.Client, error)
}
type SmtpClientPool struct {
	// pool [](*smtp.Client)
}

func (s SmtpClientPool) acquire(smtpHost, smtpPort string) (*smtp.Client, error) {
	// just create a new one for now
	if client, err := newClient(smtpHost, smtpPort); err != nil {
		return nil, err
	} else {
		return client, nil
	}
}
