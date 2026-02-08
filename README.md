# Development

## Mock SMTP server

Run mock smtp server (smtp4dev)

```bash
docker compose up
```

You can check emails in smtp4dev server at http://localhost:4999 (if you use smtp4dev)

### Alternatively, you can use mock smtp which is built with github.com/emersion/go-smtp

```bash
go run ./cmd/mocksmtp/main.go
```

## Simple email sender

Run go app to send email

```bash
go run ./cmd/emailsender/main.go
```

## Pool vs No Pool

### Pool V1

```bash
IS_POOL=true POOL_VERSION=V1 go run ./cmd/poolvsnopool/main.go
```

### Pool V2

```bash
IS_POOL=true POOL_VERSION=V2 go run ./cmd/poolvsnopool/main.go
```

### No pool V1

```bash
IS_POOL=false NO_POOL_VERSION=V1 go run ./cmd/poolvsnopool/main.go
```

### No pool V2

```bash
IS_POOL=false NO_POOL_VERSION=V2 go run ./cmd/poolvsnopool/main.go
```

### Benchmark Results

Results from sending 10000 emails.

| Strategy | Description | Time (ms) | Rate (emails/sec) |
| --- | --- | --- | --- |
| Pool V1 | Keeps smtp clients in channel | 326 | 30624.65 |
| Pool V2 | Keeps smtp clients in slice | 519 | 19266.80 |
| No Pool V1 | Create and use a single smtp.Client | 780 | 12811.71 |
| No Pool V2 | Use smtp.SendMail | 5663 | 1765.78 |

**Summary:**
- Pool is faster than no pool.
- When using pool, keeping smtp clients in channel is faster than keeping them in slice. (58.95% faster)
- When using no pool, creating a new smtp client is faster than using `smtp.SendMail`. (625.6% faster)

# Environment variables

Please check in each cmd's main.go file.
