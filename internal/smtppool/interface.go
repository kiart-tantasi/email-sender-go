package smtppool

import "net/smtp"

type ISMTPPool interface {
	Get() (*smtp.Client, error)
	Return(*smtp.Client)
}
