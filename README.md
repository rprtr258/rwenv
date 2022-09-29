# run with env file

I did not find convenient way to run command with environment from one or several env files (e.g. with db vars, mongo vars, app vars, etc.) in form of
```env
DB_HOST=db
DB_PORT=1337

# commentary
ENVIRONMENT=dev
```
so I wrote my own.

## Install
```bash
go install github.com/rprtr258/rwenv/cmd@latest
```

## Dev
You can test `rwenv` using env pretty printer like so:
```bash
go run cmd/main.go -e .env -vi go run cmd/main.go
```
