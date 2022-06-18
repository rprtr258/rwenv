run:
	go run cmd/rwenv/main.go -e .env -vi go run cmd/show_env/main.go

build:
	go build -o rwenv cmd/rwenv/main.go
