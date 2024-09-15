.PHONY:

build: 
	go build -o ./app cmd/main.go

run: build
	./app

test-auth:
	go test ./internal/auth

test-mail:
	go test ./internal/mailer
