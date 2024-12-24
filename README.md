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
go install github.com/rprtr258/rwenv@latest
```

## Usage
Here, `run_command` is command you want to run with env vars from `.env` file, e.g. `go run main.go`.
```bash
# basic usage
rwenv -ie .env run_command
# advanced usage
rwenv -vi -e .defaults.env -e .local.env -o DB_DSN=<prod> run_command
```

`rwenv` can also print resulting env by providing no command. In this case it just prints list of all env vars, bit nicer than `env`.

### Help
```bash
$ rwenv --help
NAME:
   rwenv - Run command with environment taken from file

USAGE:
   Run cmd using env:
     rwenv [-i] [-v] [-e env-file]... [-o VAR=override]... cmd...
   Show env to be used:
     rwenv [-i] [-v] [-e env-file]... [-o VAR=override]... cmd...

   Example:
     rwenv             # show env
     rwenv -e .env env # show env with added vars from .env

GLOBAL OPTIONS:
   --env value, -e value [ --env value, -e value ]            env files to take vars from
   --override value, -o value [ --override value, -o value ]  additional env vars in form of VAR_NAME=VALUE
   --verbose, -v                                              print var reading info (default: false)
   --inherit, -i                                              inherit shell env vars (default: false)
   --help, -h                                                 show help
```

## Dev
You can test `rwenv` using env pretty printer like so:
```bash
go run main.go -e .env -vi go run main.go
```
