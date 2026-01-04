package smtppool

import (
	"log"
	"net/smtp"
)

func NewPool(size int, smtpHost, smtpPort string, poolVersion string) (SMTPPool, error) {
	if poolVersion == "V2" {
		log.Println("Creating pool V2")
		return newSMTPPoolV2(size, smtpHost, smtpPort)
	}
	log.Println("Creating pool V1")
	return newSMTPPoolV1(size, smtpHost, smtpPort)
}

func NewClient(addr, helo string) (*smtp.Client, error) {
	client, err := smtp.Dial(addr)
	if err != nil {
		return nil, err
	}

	// HELO (optional)
	if helo != "" {
		if err := client.Hello(helo); err != nil {
			return nil, err
		}
	}

	return client, nil
}
