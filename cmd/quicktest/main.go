package main

import (
	"fmt"
	"log"

	"github.com/kiart-tantasi/email-sender-go/internal/smtppool"
)

// for simple and quick test

func main() {
	client, err := smtppool.NewClient("localhost:25", "")
	if err != nil {
		log.Fatal(err)
	}

	// helo only once
	client.Hello("TEST")

	for i := range 10 {
		// from
		client.Mail(fmt.Sprintf("from%d@test.com", i))
		// to
		client.Rcpt(fmt.Sprintf("to%d@test.com", i))

		// data
		writer, err := client.Data()
		if err != nil {
			log.Fatal(err)
		}
		msg := fmt.Sprintf("Subject: Hello %d\n\nThis is body", i)
		if _, err := writer.Write([]byte(msg)); err != nil {
			log.Fatal(err)
		}
		writer.Close()
	}

	log.Println("App finished")
}
