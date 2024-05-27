all: compile

compile:
	echo "Compiling for every OS and Platform"

	cd cmd/xkcd && GOOS=linux GOARCH=amd64 go build -o ../../xkcd.elf .
	cd cmd/xkcd && GOOS=windows GOARCH=amd64 go build -o ../../xkcd.exe .

run:
	@read -p "Enter a flag (optional):" flag; \
	./xkcd.elf $$flag

docker_up:
	@echo Starting Docker images..
	docker-compose up -d 
	@echo Docker images started

run_migrations:
	go install github.com/pressly/goose/v3/cmd/goose@latest
	cd internal/schema && ${GOOSE_UP}

down_migrations: 
	cd internal/schema && ${GOOSE_DOWN}

up: compile run

test: 
	go test -race -cover ./... 

## installs tools for linting
tools:
	go install github.com/daixiang0/gci@latest
	go install mvdan.cc/gofumpt@latest

## lint: runs golangci-lint on the app
lint:
	go mod tidy
	gofumpt -w .
	gci write . --skip-generated -s standard -s default
	golangci-lint run ./...

## runs lint and tool install
run_lint: tools lint

first_run: docker_up run_migrations compile run

# go test -coverprofile cover
# go tool cover -html=cover -o coverage.html