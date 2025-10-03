package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kiart-tantasi/email-sender-go/internal/env"
	"github.com/kiart-tantasi/email-sender-go/internal/smtppool"
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
var GOROUTINE_COUNT = 10
var CHANNEL_CAPACITY int = 100
var SMTP_POOL_SIZE int = 10

// CONUTERS
var sendCounter int32 = 0
var successfulSendCounter int32 = 0

func main() {
	smtpHost := env.GetEnv("SMTP_HOST", "localhost")
	smtpPort := env.GetEnv("SMTP_PORT", "25")
	emailCountStr := env.GetEnv("EMAIL_COUNT", "100")
	emailCount, err := strconv.Atoi(emailCountStr)
	if err != nil {
		log.Fatal(err)
	}

	queue := make(chan Queue, CHANNEL_CAPACITY)
	var wg sync.WaitGroup
	start := time.Now()

	// enqueue mocks
	wg.Add(1)
	go func() {
		defer wg.Done()
		from, to, msg := mockEmailPaylod()
		for range emailCount {
			queue <- Queue{
				from,
				to,
				msg,
			}
			wg.Add(1)
		}
	}()

	pool, err := smtppool.NewSMTPPool(SMTP_POOL_SIZE, smtpHost, smtpPort)
	if err != nil {
		log.Fatalf("Error while creating smtp pool: %v", err)
	}

	// create multiple goroutines to send email to smtp server
	for i := range GOROUTINE_COUNT {
		go func(idx int) {
			log.Printf("Started goroutine %d", idx)

			// infinite loop
			for {
				select {
				case q := <-queue:

					client, err := pool.Get()
					if err != nil {
						log.Printf("Errror when getting smtp client: %s", err)
						return
					}

					client.Hello("TEST")
					client.Mail(q.from)
					client.Rcpt(q.to[0])
					writer, err := client.Data()
					// reconnect if client gets error
					if err != nil {
						log.Printf("Client data error, attempting to reconnect: %v", err)
						client, err = smtppool.NewSMTPClient(fmt.Sprintf("%s:%s", smtpHost, smtpPort))
						if err != nil {
							log.Printf("Error when creating new smtp client: %s", err)
							pool.Return(client)
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

					// Stats
					atomic.AddInt32(&sendCounter, 1)
					// Return smtp client to pool
					pool.Return(client)
					wg.Done()
				case <-time.After(10 * time.Second):
					log.Fatalln("Error: No message received for 10 seconds so app wil exit")
				}
			}

		}(i)
	}

	// wait
	wg.Wait()
	fmt.Printf("%d send, %d goroutines, %d smtp clients, %d channel capacity, %d ms, (successful send: %d)\n", sendCounter, GOROUTINE_COUNT, SMTP_POOL_SIZE, CHANNEL_CAPACITY, time.Since(start).Milliseconds(), successfulSendCounter)
}

func mockEmailPaylod() (string, []string, []byte) {
	from := "from@test.com"
	to := []string{"to@test.com"}
	msg := "Subject: Hello from Example.com !\n\nThis is body"
	return from, to, []byte(msg)
}
