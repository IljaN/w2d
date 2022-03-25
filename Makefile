build_all: bin/w2d-amd64-linux bin/w2d-amd64-darwin bin/w2d-amd64-windows.exe

bin/w2d-amd64-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/w2d-amd64-linux main.go

bin/w2d-amd64-darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o bin/w2d-amd64-darwin main.go

bin/w2d-amd64-windows.exe:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/w2d-amd64-windows.exe main.go

.PHONY: test
test:
	go test -v ./...

.PHONY: test-coverage
coverage:
	go test -coverprofile cover.out -v ./...

.PHONY: show-coverage
show-coverage: coverage
	go tool cover -html=cover.out

.PHONY: clean
clean:
	rm -rf ./bin w2d cover.out
