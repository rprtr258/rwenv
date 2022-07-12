run:
	go run cmd/rwenv/main.go -vie .env go run cmd/show_env/main.go

run2:
	go run cmd/rwenv/main.go -vi go run cmd/rwenv/main.go -e .env -vi go run cmd/show_env/main.go

build:
	go build -o rwenv cmd/rwenv/main.go
