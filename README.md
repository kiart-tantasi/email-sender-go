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

### Pool v1

```bash
IS_POOL=true POOL_VERSION=V1 go run ./cmd/poolvsnopool/main.go
```

### Pool v2

```bash
IS_POOL=true POOL_VERSION=V2 go run ./cmd/poolvsnopool/main.go
```

### No pool v1

```bash
IS_POOL=false NO_POOL_VERSION=V1 go run ./cmd/poolvsnopool/main.go
```

### No pool v2

```bash
IS_POOL=false NO_POOL_VERSION=V2 go run ./cmd/poolvsnopool/main.go
```

# Config/Env vars

There are some env vars. Please check in main.go.
