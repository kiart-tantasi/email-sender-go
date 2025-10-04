package smtppool

import (
	"log"
	"net/smtp"

	"github.com/kiart-tantasi/email-sender-go/internal/env"
)

func NewPool(size int, smtpHost, smtpPort string) (ISMTPPool, error) {
	if env.GetEnv("POOL_VERSION", "V1") == "V2" {
		log.Println("Creating pool v2")
		return newSMTPPoolV2(size, smtpHost, smtpPort)
	}
	log.Println("Creating pool v1")
	return newSMTPPoolV1(size, smtpHost, smtpPort)
}

func NewClient(addr string) (*smtp.Client, error) {
	client, err := smtp.Dial(addr)
	if err != nil {
		return nil, err
	}
	return client, nil
}
