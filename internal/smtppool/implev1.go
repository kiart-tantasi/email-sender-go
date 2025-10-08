package smtppool

import (
	"fmt"
	"log"
	"net/smtp"
)

type SMTPPoolV1 struct {
	clients chan *smtp.Client
	addr    string
}

func newSMTPPoolV1(size int, smtpHost, smtpPort string) (SMTPPool, error) {
	clients := make(chan *smtp.Client, size)
	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)

	for range size {
		if client, err := NewClient(addr, ""); err != nil {
			return nil, err
		} else {
			clients <- client
		}
	}

	return &SMTPPoolV1{clients: clients, addr: addr}, nil
}

func (p *SMTPPoolV1) Get() (*smtp.Client, error) {
	client, ok := <-p.clients
	if !ok {
		return nil, fmt.Errorf("smtp client channel is closed")
	}
	return client, nil
}

func (p *SMTPPoolV1) Return(client *smtp.Client) {
	select {
	case p.clients <- client:
		// do nothing
	default:
		client.Close()
		log.Println("Channel is full. Closing connection...")
	}
}
