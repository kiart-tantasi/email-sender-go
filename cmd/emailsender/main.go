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

/*
[RESULTS]

(macbook air m2)

# 100

100 sent, 1 goroutines, - smtp  clients, 100 channel capacity, 476 ms
100 sent, 1 goroutines, 1 smtp  clients, 100 channel capacity, 297 ms
100 sent, 10 goroutines, - smtp  clients, 100 channel capacity, 343 ms
100 sent, 10 goroutines, 10 smtp  clients, 100 channel capacity, 180 ms
100 sent, 100 goroutines, - smtp  clients, 100 channel capacity, 271 ms
100 sent, 100 goroutines, 100 smtp  clients, 100 channel capacity, 243 ms

# 1,000

1000 sent, 1 goroutines, - smtp clients, 100 channel capacity, 4760 ms
1000 sent, 1 goroutines, 1 smtp clients, 100 channel capacity, 2255 ms
1000 sent, 10 goroutines, - smtp clients, 100 channel capacity, 2981 ms
1000 sent, 10 goroutines, 10 smtp clients, 100 channel capacity, 1571 ms
1000 sent, 100 goroutines, - smtp clients, 100 channel capacity, 2919 ms
1000 sent, 100 goroutines, 100 smtp clients, 100 channel capacity, 1601 ms

# 10,000

10000 sent, 1 goroutines, - smtp clients, 100 channel capacity, 66930 ms
10000 sent, 1 goroutines, 1 smtp clients, 100 channel capacity, 20617 ms
10000 sent, 10 goroutines, - smtp clients, 100 channel capacity, 46406 ms
10000 sent, 10 goroutines, 10 smtp clients, 100 channel capacity, 15192 ms
10000 sent, 100 goroutines, - smtp clients, 100 channel capacity, 44224 ms
10000 sent, 100 goroutines, 100 smtp clients, 100 channel capacity, 17212 ms

# 100,000

100000 sent, 1 goroutines, 1 smtp clients, 100 channel capacity, 195919 ms
100000 sent, 10 goroutines, 10 smtp  clients, 100 channel capacity, 157515 ms
100000 sent, 100 goroutines, 100 smtp clients, 100 channel capacity, 161958 ms

# 10,000 (2025-10-05)

10000 sent, 100 goroutines, 10 smtp clients, 100 channel capacity, XXX ms
10000 sent, 100 goroutines, 100 smtp clients, 100 channel capacity, XXX ms
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
var countSent int32 = 0

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

	pool, err := smtppool.NewPool(SMTP_POOL_SIZE, smtpHost, smtpPort)
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
						client, err = smtppool.NewClient(fmt.Sprintf("%s:%s", smtpHost, smtpPort))
						if err != nil {
							log.Printf("Error when creating new smtp client: %s", err)
							pool.Return(client)
							return
						}
					}

					// Send email
					if _, err := writer.Write(q.msg); err != nil {
						log.Printf("Error when sending email on goroutine %d: %s", i, err)
					} else {
						// Close
						writer.Close()
						// Stats
						atomic.AddInt32(&countSent, 1)
					}

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
	fmt.Printf("%d sent, %d goroutines, %d smtp clients, %d channel capacity, %d ms\n", countSent, GOROUTINE_COUNT, SMTP_POOL_SIZE, CHANNEL_CAPACITY, time.Since(start).Milliseconds())
}

func mockEmailPaylod() (string, []string, []byte) {
	from := "from@test.com"
	to := []string{"to@test.com"}
	msg := "Subject: Hello from Example.com !\n\nThis is body"
	return from, to, []byte(msg)
}
