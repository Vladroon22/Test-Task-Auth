FROM golang:1.23.1-alpine AS builder

WORKDIR /usr/local/src

RUN apk --no-cache add git make

COPY ["go.mod", "go.sum", "./"]

RUN go mod download

COPY . ./

RUN go build -o app ./cmd/main.go



FROM alpine

COPY --from=builder /usr/local/src/app /
COPY ./config/config.go ./config.go

ENV DB="postgres:11111@localhost:5431/postgres?sslmode=disable"

CMD [ "./app" ]

EXPOSE 3000

