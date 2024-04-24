all: compile

compile:
	echo "Compiling for every OS and Platform"

	cd cmd/xkcd && GOOS=linux GOARCH=amd64 go build -o ../../xkcd.elf .
	cd cmd/xkcd && GOOS=windows GOARCH=amd64 go build -o ../../xkcd.exe .

run:
	@read -p "Enter a flag (optional):" flag; \
	./xkcd.elf $$flag

up: compile run
