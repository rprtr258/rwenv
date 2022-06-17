# run with env file

I did not find convenient way to run command with environment from one or several env files (e.g. with db vars, mongo vars, app vars, etc.) in form of
```
DB_HOST=db
DB_PORT=1337

# commentary
ENVIRONMENT=dev
```
so I wrote my own.

## Install
```bash
go install github.com/rprtr258/rwenv/cmd/rwenv@latest
```

## Usage
```bash
rwenv -e .env env
```