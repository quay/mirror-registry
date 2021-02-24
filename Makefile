include .env

all:

full-reset: .
	go build main.go; sudo ./main uninstall; sudo ./main install