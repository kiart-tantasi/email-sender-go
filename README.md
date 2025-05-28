# Development

Run fake smtp server
```
docker compose up
```

Run go app to send email
```
go run ./apps/emailsender/cmd/emailsender/main.go

# with env vars
SMTP_HOST=? SMTP_PORT=? SMTP_USERNAME=? SMTP_PASSWORD=? go run apps/emailsender/cmd/emailsender/main.go

# no auth
SMTP_HOST=? SMTP_PORT=? go run apps/emailsender/cmd/emailsender/main.go
```

Check emails in smtp4dev server at http://localhost:4999
