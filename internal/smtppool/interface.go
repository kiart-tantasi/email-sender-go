package smtppool

import "net/smtp"

type SMTPPool interface {
	Get() (*smtp.Client, error)
	Return(*smtp.Client)
}
