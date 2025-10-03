package smtppool

import (
	"fmt"
	"net/smtp"
)

type SMTPPool struct {
	clients chan *smtp.Client
}

func NewSMTPPool(size int, smtpHost, smtpPort string) (*SMTPPool, error) {

	clients := make(chan *smtp.Client, size)

	for range size {
		if client, err := NewSMTPClient(fmt.Sprintf("%s:%s", smtpHost, smtpPort)); err != nil {
			return nil, err
		} else {
			clients <- client
		}
	}

	return &SMTPPool{clients: clients}, nil
}

func (p *SMTPPool) Get() (*smtp.Client, error) {
	client, ok := <-p.clients
	if !ok {
		return nil, fmt.Errorf("smtp client channel is closed")
	}
	return client, nil
}

func (p *SMTPPool) Keep(client *smtp.Client) {
	p.clients <- client
}

func NewSMTPClient(addr string) (*smtp.Client, error) {
	client, err := smtp.Dial(addr)
	if err != nil {
		return nil, err
	}
	return client, nil
}
