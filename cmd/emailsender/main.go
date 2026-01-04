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

# pool V1, 10000 sent, gosmtp

10000 sent, 100 goroutines, 1 smtp clients, 100 channel capacity, 878 ms
10000 sent, 100 goroutines, 2 smtp clients, 100 channel capacity, 508 ms
10000 sent, 100 goroutines, 5 smtp clients, 100 channel capacity, 443 ms
10000 sent, 100 goroutines, 10 smtp clients, 100 channel capacity, 331 ms
10000 sent, 100 goroutines, 50 smtp clients, 100 channel capacity, 249 ms
10000 sent, 100 goroutines, 100 smtp clients, 100 channel capacity, 247 ms

summary: after 50 smtp clients, there is no increase on performance.

# pool V2, 10000sent, gosmtp

10000 sent, 100 goroutines, 1 smtp clients, 100 channel capacity, 1069 ms
10000 sent, 100 goroutines, 2 smtp clients, 100 channel capacity, 592 ms
10000 sent, 100 goroutines, 5 smtp clients, 100 channel capacity, 553 ms
10000 sent, 100 goroutines, 10 smtp clients, 100 channel capacity, 511 ms
10000 sent, 100 goroutines, 50 smtp clients, 100 channel capacity, 508 ms
10000 sent, 100 goroutines, 100 smtp clients, 100 channel capacity, 507 ms

*/

type Queue struct {
	from string
	to   []string
	msg  []byte
}

func main() {
	// vars
	goroutineCountStr := env.GetEnv("GOROUTINE_COUNT", "100")
	smtpPoolSizeStr := env.GetEnv("POOL_SIZE", "10")
	channelCapacityStr := env.GetEnv("CHANNEL_CAPACITY", "100")
	smtpHost := env.GetEnv("SMTP_HOST", "localhost")
	smtpPort := env.GetEnv("SMTP_PORT", "25")
	emailCountStr := env.GetEnv("EMAIL_COUNT", "100")
	poolVersion := env.GetEnv("POOL_VERSION", "V1")

	// cast string to int
	goroutineCount, err := strconv.Atoi(goroutineCountStr)
	if err != nil {
		log.Fatal(err)
	}
	smtpPoolSize, err := strconv.Atoi(smtpPoolSizeStr)
	if err != nil {
		log.Fatal(err)
	}
	channelCapacity, err := strconv.Atoi(channelCapacityStr)
	if err != nil {
		log.Fatal(err)
	}
	emailCount, err := strconv.Atoi(emailCountStr)
	if err != nil {
		log.Fatal(err)
	}

	// stats
	var countSent int32 = 0

	queue := make(chan Queue, channelCapacity)
	var wg sync.WaitGroup

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

	pool, err := smtppool.NewPool(smtpPoolSize, smtpHost, smtpPort, poolVersion)
	if err != nil {
		log.Fatalf("Error while creating smtp pool: %v", err)
	}

	// excluding pool init time
	start := time.Now()

	// create multiple goroutines to send email to smtp server
	for i := range goroutineCount {
		go func(idx int) {
			for {
				select {
				case q := <-queue:

					client, err := pool.Get()
					if err != nil {
						log.Printf("Errror when getting smtp client: %s", err)
						return
					}

					client.Mail(q.from)
					client.Rcpt(q.to[0])
					writer, err := client.Data()
					// reconnect if client gets error
					if err != nil {
						log.Printf("Client data error, attempting to reconnect: %v", err)
						client, err = smtppool.NewClient(fmt.Sprintf("%s:%s", smtpHost, smtpPort), "")
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
	fmt.Printf("%d sent, %d goroutines, %d smtp clients, %d channel capacity, %d ms\n", countSent, goroutineCount, smtpPoolSize, channelCapacity, time.Since(start).Milliseconds())
}

func mockEmailPaylod() (string, []string, []byte) {
	from := "from@test.com"
	to := []string{"to@test.com"}
	msg := "Subject: Hello from Example.com !\n\nThis is body"
	return from, to, []byte(msg)
}
