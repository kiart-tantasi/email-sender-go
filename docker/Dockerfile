from golang:1.25

COPY ./apps/emailsender /emailsender

WORKDIR /emailsender

RUN go build -o app ./cmd/mock/main.go

CMD ./app
