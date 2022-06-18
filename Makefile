run:
	go run cmd/rwenv/main.go -e .env -vi go run cmd/show_env/main.go

install: uninstall
	go install github.com/rprtr258/rwenv/cmd/rwenv@latest

uninstall:
	rm ${}/rwenv || true
