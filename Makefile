.PHONY: run build

run:
	go run cmd/api/main.go

dev:
	air

build:
	go build -o api.exe cmd/api/main.go
