BINARY_PATH = build/rvm-backend
MAIN_FILE = cmd/backend/main.go

.PHONY: all run/dev build build/prod clean dep vet tidy lint

run/dev: build
	APP_ENV=development ./${BINARY_PATH}

build:
	go build -o ${BINARY_PATH} ${MAIN_FILE}

build/prod:
	GOOS=linux GOARCH=amd64 go build -o ${BINARY_PATH}-linux ${MAIN_FILE}
	GOOS=darwin GOARCH=amd64 go build -o ${BINARY_PATH}-darwin ${MAIN_FILE}
	GOOS=windows GOARCH=amd64 go build -o ${BINARY_PATH}-windows ${MAIN_FILE}

clean:
	go clean
	rm -f ${BINARY_PATH}
	rm -f ${BINARY_PATH}-darwin
	rm -f ${BINARY_PATH}-linux
	rm -f ${BINARY_PATH}-windows

dep:
	go mod download

vet:
	go vet

tidy:
	go fmt ./...
	go mod tidy -v

lint:
	golangci-lint run
