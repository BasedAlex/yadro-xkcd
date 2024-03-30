all: compile

run:
	@read -p "Enter a string:" flag; \
	go run . -s "$$flag"

compile:
	echo "Compiling for every OS and Platform"
	GOOS=linux GOARCH=amd64 go build -o bin/main-linux-arm64 .
	GOOS=windows GOARCH=amd64 go build -o bin/win .

up:
	@read -p "Please enter a string:" flag; \
	bin/main-linux-arm64/yadro-xkcd -s "$$flag"

up_windows:
	@read -p "Please enter a string:" flag; \
	bin/win/yadro-xkcd.exe -s "$$flag"
