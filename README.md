# Development

Run fake smtp server

```bash
docker compose up
```

Alternatively, you can use mock smtp which is built with github.com/emersion/go-smtp

```bash
go run ./cmd/mocksmtp/main.go
```

Run go app to send email

```bash
go run ./cmd/emailsender/main.go
```

There are some env vars. Please check in main.go.

Check emails in smtp4dev server at http://localhost:4999 (if you use smtp4dev)
