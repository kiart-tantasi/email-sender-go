package smtppool

import (
	"fmt"
	"log"
	"net/smtp"
	"sync"

	"github.com/kiart-tantasi/email-sender-go/internal/env"
)

// NOTE: pool v2 has issue that it creates more connections more than pool size.
// The cause is method "Get" always create a new smtp client instead of waiting for ones in progress.

type SMTPPoolV2 struct {
	clients []*smtp.Client
	mu      sync.Mutex
	addr    string
	size    int
}

func newSMTPPoolV2(size int, smtpHost, smtpPort string) (SMTPPool, error) {
	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
	clients := []*smtp.Client{}

	// preload connections
	if env.GetEnv("SMTP_PRELOAD", "true") == "true" {
		log.Printf("Preloading %d connections...", size)
		for range size {
			client, err := NewClient(addr, "")
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
		client.Close()
	}

	// When smtp clients are available, create new smtp client
	return NewClient(p.addr, "")
}

func (p *SMTPPoolV2) Return(client *smtp.Client) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Pool is full. Close and discard connection
	if len(p.clients) > p.size {
		client.Close()
		return
	}
	p.clients = append(p.clients, client)
}
