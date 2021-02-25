include .env

all:

full-reset: .
	go build main.go; sudo ./main uninstall -v; sudo ./main install -v