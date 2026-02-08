package main

import (
	"fmt"
	"log"
	"net/smtp"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kiart-tantasi/email-sender-go/internal/env"
	"github.com/kiart-tantasi/email-sender-go/internal/smtppool"
)

// NOTE: results are in README.md

func main() {
	// env vars
	isPool := env.GetEnv("IS_POOL", "true") == "true"

	smtpHost := env.GetEnv("SMTP_HOST", "localhost")
	smtpPort := env.GetEnv("SMTP_PORT", "25")
	emailCountStr := env.GetEnv("EMAIL_COUNT", "10000")
	noPoolVersion := env.GetEnv("NO_POOL_VERSION", "V1")
	poolVersion := env.GetEnv("POOL_VERSION", "V1")
	semLimit := 10

	// cast
	emailCount, err := strconv.Atoi(emailCountStr)
	if err != nil {
		log.Fatal(err)
	}

	var countSent int64 = 0
	start := time.Now()

	// send emails
	if isPool {
		if poolVersion == "V1" {
			log.Println("Pool V1")
			PoolV1(smtpHost, smtpPort, emailCount, &countSent, semLimit)
		} else {
			log.Println("Pool V2")
			PoolV2(smtpHost, smtpPort, emailCount, &countSent, semLimit)
		}
	} else {
		if noPoolVersion == "V1" {
			log.Println("No pool V1")
			NoPoolV1(smtpHost, smtpPort, emailCount, &countSent, semLimit)
		} else {
			log.Println("No pool V2")
			NoPoolV2(smtpHost, smtpPort, emailCount, &countSent, semLimit)
		}
	}

	taken := time.Since(start)
	log.Printf("Sent %d emails in %d ms (%f emails/sec)", countSent, taken.Milliseconds(), (float64(countSent) / taken.Seconds()))
}

// Keeps smtp clients in channel
func PoolV1(smtpHost, smtpPort string, emailCount int, countSent *int64, semLimit int) {
	Pool(smtpHost, smtpPort, emailCount, countSent, "V1", semLimit)
}

// Keeps smtp clients in slice
func PoolV2(smtpHost, smtpPort string, emailCount int, countSent *int64, semLimit int) {
	Pool(smtpHost, smtpPort, emailCount, countSent, "V2", semLimit)
}

func Pool(smtpHost, smtpPort string, emailCount int, countSent *int64, poolVersion string, semLimit int) {
	pool, err := smtppool.NewPool(10, smtpHost, smtpPort, poolVersion)
	if err != nil {
		log.Fatal(err)
	}

	sem := make(chan any, semLimit)
	var wg sync.WaitGroup

	for i := range emailCount {
		sem <- struct{}{}
		wg.Add(1)
		go func() {
			defer func() {
				<-sem
				wg.Done()
			}()

			client, err := pool.Get()
			if err != nil {
				log.Fatal(err)
			}
			// MAIL FROM
			client.Mail(fmt.Sprintf("from%d@test.com", i))
			// RCPT TO
			client.Rcpt(fmt.Sprintf("to%d@test.com", i))
			// DATA
			writer, err := client.Data()
			// reconnect if client gets error
			if err != nil {
				log.Printf("Client data error, attempting to reconnect: %v", err)
				client, err = smtppool.NewClient(fmt.Sprintf("%s:%s", smtpHost, smtpPort), "")
				if err != nil {
					log.Fatalf("Error when creating new smtp client: %s", err)
				}
			}
			_, err = writer.Write([]byte(fmt.Sprintf("subject: FOOBAR%d\n\ntest body\n", i)))
			if err != nil {
				log.Fatal(err)
			}

			writer.Close()
			pool.Return(client)
			atomic.AddInt64(countSent, 1)
		}()
	}

	wg.Wait()
}

// Create and use a single smtp.Client (goroutines cannot be used here because 1 smtp client can handle 1 smtp request at a time)
func NoPoolV1(smtpHost, smtpPort string, emailCount int, countSent *int64, semLimit int) {
	client, err := smtppool.NewClient(fmt.Sprintf("%s:%s", smtpHost, smtpPort), "")
	if err != nil {
		log.Fatalf("Error when creating smtp client: %v", err)
	}

	for i := range emailCount {
		// MAIL FROM
		client.Mail(fmt.Sprintf("from%d@test.com", i))
		// RCPT TO
		client.Rcpt(fmt.Sprintf("to%d@test.com", i))
		// DATA
		writer, err := client.Data()
		// reconnect if client gets error
		if err != nil {
			log.Printf("Client data error, attempting to reconnect: %v", err)
			client, err = smtppool.NewClient(fmt.Sprintf("%s:%s", smtpHost, smtpPort), "")
			if err != nil {
				log.Fatalf("Error when creating new smtp client: %s", err)
			}
		}
		_, err = writer.Write([]byte(fmt.Sprintf("subject: FOOBAR%d\n\ntest body\n", i)))
		if err != nil {
			log.Fatal(err)
		}

		writer.Close()
		atomic.AddInt64(countSent, 1)
	}
}

// Use smtp.SendMail
func NoPoolV2(smtpHost, smtpPort string, emailCount int, countSent *int64, semLimit int) {
	sem := make(chan any, semLimit)
	var wg sync.WaitGroup

	for i := range emailCount {
		sem <- struct{}{}
		wg.Add(1)
		go func() {
			defer func() {
				<-sem
				wg.Done()
			}()

			err := smtp.SendMail(
				fmt.Sprintf("%s:%s", smtpHost, smtpPort),
				nil,
				fmt.Sprintf("from%d@test.com", i),
				[]string{fmt.Sprintf("to%d@test.com", i)}, []byte(fmt.Sprintf("subject: FOOBAR%d\n\ntest body\n", i)),
			)
			if err != nil {
				log.Fatalf("Error when sending email: %v", err)
			}
			atomic.AddInt64(countSent, 1)
		}()
	}

	wg.Wait()
}
