include .env

all:

full-reset: .
	go build main.go; sudo ./main uninstall -v; sudo ./main install -v

build-online-zip:
	sudo podman run --rm -v ${PWD}:/usr/src:Z -w /usr/src docker.io/golang:1.16 go build -v -o quay-installer;
	tar -cvzf quay-installer.tar.gz quay-installer README.md
	rm -f quay-installer

build-offline-zip:
	sudo podman run --rm -v ${PWD}:/usr/src:Z -w /usr/src docker.io/golang:1.16 go build -v -o quay-installer;
	sudo podman save \
	--multi-image-archive \
	docker.io/centos/postgresql-10-centos8 \
	quay.io/projectquay/quay:latest \
	docker.io/centos/redis-5-centos8 \
	> image-archive.tar
	tar -cvzf quay-installer.tar.gz quay-installer README.md image-archive.tar
	rm -rf quay-installer image-archive.tar

build-image-archive: 
	sudo podman save \
	--multi-image-archive \
	docker.io/centos/postgresql-10-centos8 \
	quay.io/projectquay/quay:latest \
	docker.io/centos/redis-5-centos8 \
	> image-archive.tar
	
clean:
	rm -rf quay-installer* image-archive.tar