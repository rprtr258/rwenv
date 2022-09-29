RWENV:="go run main.go"

# test run on .env file
run-dotenv:
	{{RWENV}} -vie .env {{RWENV}}

# test run rwenv on rwenv on .env file
run-rwenv-dotenv:
	{{RWENV}} -vi {{RWENV}} -e .env -vi {{RWENV}}

# run linter
@lint:
	golangci-lint run --exclude-use-default=false --disable-all \
		--enable=revive --enable=deadcode --enable=errcheck --enable=govet --enable=ineffassign --enable=structcheck --enable=typecheck --enable=varcheck --enable=asciicheck --enable=bidichk --enable=bodyclose --enable=containedctx --enable=contextcheck --enable=cyclop --enable=decorder --enable=depguard --enable=dogsled --enable=dupl --enable=durationcheck --enable=errchkjson --enable=errname --enable=errorlint --enable=execinquery --enable=exhaustive --enable=exhaustruct --enable=exportloopref --enable=forbidigo --enable=forcetypeassert --enable=funlen --enable=gochecknoinits --enable=gocognit --enable=goconst --enable=gocritic --enable=gocyclo --enable=godot --enable=godox --enable=goerr113 --enable=gofmt --enable=gofumpt --enable=goimports --enable=gomnd \
		--enable=gomoddirectives --enable=gomodguard --enable=goprintffuncname --enable=gosec --enable=grouper --enable=ifshort --enable=importas --enable=lll --enable=maintidx --enable=makezero --enable=misspell --enable=nestif --enable=nilerr --enable=nilnil --enable=noctx --enable=nolintlint --enable=nosprintfhostport --enable=paralleltest --enable=prealloc --enable=predeclared --enable=promlinter --enable=rowserrcheck --enable=sqlclosecheck --enable=tenv --enable=testpackage --enable=thelper --enable=tparallel --enable=unconvert --enable=unparam --enable=wastedassign --enable=whitespace --enable=wrapcheck

# show list of all todos left in code
@todo:
	rg 'TODO' --glob '**/*.go' || echo 'All done!'

# build executable
@build:
	go build -o rwenv main.go
