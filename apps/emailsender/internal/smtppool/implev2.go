package smtppool

import (
	"fmt"
	"log"
	"net/smtp"
	"sync"

	"github.com/kiart-tantasi/email-sender-go/internal/env"
)

type SMTPPoolV2 struct {
	clients []*smtp.Client
	mu      sync.Mutex
	addr    string
	size    int
}

func newSMTPPoolV2(size int, smtpHost, smtpPort string) (ISMTPPool, error) {
	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
	clients := []*smtp.Client{}

	// preload connections
	if env.GetEnv("SMTP_PRELOAD", "true") == "true" {
		log.Printf("Preloading %d connections...", size)
		for range size {
			client, err := NewClient(addr)
			if err != nil {
				return nil, err
			}
			clients = append(clients, client)
		}
	}

	return &SMTPPoolV2{
			clients: clients,
			addr:    addr,
			size:    size},
		nil
}

func (p *SMTPPoolV2) Get() (*smtp.Client, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// check existing clients
	i := 0
	for len(p.clients) > 0 {
		client := p.clients[i]
		p.clients = p.clients[1:]
		err := client.Noop()
		// Okay
		if err == nil {
			return client, nil
		}
		// Non-okay
		log.Printf("SMTP Client index %d is not okay (NOOP). Closing its connection...", i)
		client.Close()
		client = nil
	}

	log.Printf("No smtp clients are okay. Recreating new smtp client...")
	return NewClient(p.addr)
}

func (p *SMTPPoolV2) Return(client *smtp.Client) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.clients) > p.size {
		log.Println("Pool is full. Closing a smtp client connection...")
		client.Close()
		return
	}
	p.clients = append(p.clients, client)
}
