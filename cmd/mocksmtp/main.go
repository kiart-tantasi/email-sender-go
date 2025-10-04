package main

import (
	"bytes"
	"io"
	"log"
	"net/mail"
	"sync/atomic"
	"time"

	"github.com/emersion/go-smtp"
)

var counter int64 = 0

type backend struct{}

type Session struct {
	from      string
	to        []string
	startTime *time.Time
}

func (be backend) NewSession(conn *smtp.Conn) (smtp.Session, error) {
	startTime := time.Now()
	return &Session{startTime: &startTime}, nil
}

// MAIL FROM
func (s *Session) Mail(from string, opts *smtp.MailOptions) error {
	s.from = from
	return nil
}

// RCPT TO
func (s *Session) Rcpt(to string, opts *smtp.RcptOptions) error {
	s.to = append(s.to, to)
	return nil
}

// DATA
func (s *Session) Data(r io.Reader) error {
	_, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	atomic.AddInt64(&counter, 1)

	// log time taken when finish
	if s.startTime != nil {
		log.Printf("session-time: %d ms (%d total emails received)", time.Since(*s.startTime).Milliseconds(), counter)
	}

	// debug
	// log.Printf("FROM: %s", s.from)
	// log.Printf("TO: %s", s.to)
	// if err := logEmailHeaders(bs); err != nil {
	// 	log.Fatal(err)
	// }

	return nil
}

// RSET
func (s *Session) Reset() {
	s.from = ""
	s.to = nil
}

func (s *Session) Logout() error {
	return nil
}

func main() {
	be := &backend{}

	// configs
	s := smtp.NewServer(be)
	s.Addr = ":25"
	s.Domain = "localhost"
	s.ReadTimeout = 10 * time.Second
	s.WriteTimeout = 10 * time.Second
	s.MaxMessageBytes = 1024 * 1024
	s.AllowInsecureAuth = true

	// start smtp server
	log.Println("Starting SMTP server at", s.Addr)
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func logEmailHeaders(bs []byte) error {
	msg, err := mail.ReadMessage(bytes.NewReader(bs))
	if err != nil {
		return err
	}

	for k, v := range msg.Header {
		log.Printf("Email header [%s], value [%s]", k, v)
	}
	return nil
}
