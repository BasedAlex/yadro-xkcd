
build:
	go build -o cmd/cli

run:
	@read -p "Enter a flag:" flag; \
	go run cmd/cli/main.go $$flag

compile:
	echo "Compiling for every OS and Platform"
	GOOS=linux GOARCH=arm64 go build -o bin/main-linux-arm64 ./cmd/cli/
	GOOS=windows GOARCH=amd64 go build -o bin/win ./cmd/cli/

up:
	@read -p "Please enter a flag:" flag; \
	bin/main-linux-arm64/cli $$flag

up_windows:
	@read -p "Please enter a flag:" flag; \
	bin/win/cli $$flag

all: compile up
