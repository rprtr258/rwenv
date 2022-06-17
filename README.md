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
go install github.com/rprtr258/rwenv/cmd/rwenv@latest
```

## Usage
```
rwenv [flags] ...cmd

Flags:
  -e, --env strings        Env files to take vars from
  -h, --help               help for rwenv
  -i, --inherit            Inherit shell env vars
  -o, --override strings   Additional env vars in form of VAR_NAME=VALUE
  -v, --verbose            Print var reading info
```

### Examples
```bash
rwenv -e .env -o DB_HOST=test env
```

## Dev
You can test `rwenv` using env pretty printer like so:
```bash
go run cmd/rwenv/main.go -e .env -vi go run cmd/show_env/main.go
```
