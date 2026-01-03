package main

import (
	"fmt"
	"log"
	"net/smtp"
	"strconv"
	"time"

	"github.com/kiart-tantasi/email-sender-go/internal/env"
	"github.com/kiart-tantasi/email-sender-go/internal/smtppool"
)

// pool v1 - Sent 10000 emails in 850 ms (11762.241001 emails/sec)
// pool v2 - Sent 10000 emails in 1050 ms (9516.048660 emails/sec)
// no pool v1 - Sent 10000 emails in 819 ms (12198.471562 emails/sec)
// no pool v2 - Sent 10000 emails in 21985 ms (454.839723 emails/sec)
// [NOTE: all of these are not using goroutines]

func main() {
	// env vars
	isPool := env.GetEnv("IS_POOL", "true") == "true"

	smtpHost := env.GetEnv("SMTP_HOST", "localhost")
	smtpPort := env.GetEnv("SMTP_PORT", "25")
	emailCountStr := env.GetEnv("EMAIL_COUNT", "10000")
	noPoolVersion := env.GetEnv("NO_POOL_VERSION", "V1")
	poolVersion := env.GetEnv("POOL_VERSION", "V1")

	// cast
	emailCount, err := strconv.Atoi(emailCountStr)
	if err != nil {
		log.Fatal(err)
	}

	countSent := 0
	start := time.Now()

	// send emails
	if isPool {
		log.Println("Using pool")
		Pool(smtpHost, smtpPort, emailCount, &countSent, poolVersion)
	} else {
		if noPoolVersion == "V1" {
			log.Println("Not using pool V1")
			NoPoolV1(smtpHost, smtpPort, emailCount, &countSent)
		} else {
			log.Println("Not using pool V2")
			NoPoolV2(smtpHost, smtpPort, emailCount, &countSent)

		}
	}

	taken := time.Since(start)
	log.Printf("Sent %d emails in %d ms (%f emails/sec)", countSent, taken.Milliseconds(), (float64(countSent) / taken.Seconds()))
}

func Pool(smtpHost, smtpPort string, emailCount int, countSent *int, poolVersion string) {
	pool, err := smtppool.NewPool(10, smtpHost, smtpPort, poolVersion)
	if err != nil {
		log.Fatal(err)
	}

	for i := range emailCount {
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
		*countSent++
	}
}

// single smtp.Client
func NoPoolV1(smtpHost, smtpPort string, emailCount int, countSent *int) {
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
		*countSent++
	}
}

// smtp.SendMail
func NoPoolV2(smtpHost, smtpPort string, emailCount int, countSent *int) {
	for i := range emailCount {
		err := smtp.SendMail(
			fmt.Sprintf("%s:%s", smtpHost, smtpPort),
			nil,
			fmt.Sprintf("from%d@test.com", i),
			[]string{fmt.Sprintf("to%d@test.com", i)}, []byte(fmt.Sprintf("subject: FOOBAR%d\n\ntest body\n", i)))
		if err != nil {
			log.Fatalf("Error when sending email: %v", err)
		}
		*countSent++
	}
}
