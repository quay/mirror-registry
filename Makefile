include .env

all:

build-ansible-ee:
	sudo ansible-builder build --container-runtime podman --file ansible-runner/execution-environment.yml --context ansible-runner/context --tag quay.io/quay/openshift-mirror-registry-ee
	sudo podman save \
	quay.io/quay/openshift-mirror-registry-ee \
	> execution-environment.tar

build-image-bundle: 
	sudo podman pull docker.io/centos/postgresql-10-centos8
	sudo podman pull docker.io/centos/redis-5-centos8 
	sudo podman pull quay.io/projectquay/quay:latest
	sudo podman save \
	--multi-image-archive \
	docker.io/centos/postgresql-10-centos8 \
	quay.io/projectquay/quay:latest \
	docker.io/centos/redis-5-centos8 \
	> image-bundle.tar

build-online-zip: build-ansible-ee
	sudo podman run --rm -v ${PWD}:/usr/src:Z -w /usr/src docker.io/golang:1.16 go build -v -o openshift-mirror-registry;
	tar -cvzf openshift-mirror-registry.tar.gz openshift-mirror-registry README.md execution-environment.tar
	rm -f openshift-mirror-registry execution-environment.tar

build-offline-zip: build-image-bundle build-ansible-ee
	sudo podman run --rm -v ${PWD}:/usr/src:Z -w /usr/src docker.io/golang:1.16 go build -v -o openshift-mirror-registry;
	tar -cvzf openshift-mirror-registry.tar.gz openshift-mirror-registry README.md execution-environment.tar image-bundle.tar
	rm -rf openshift-mirror-registry image-archive.tar execution-environment.tar

	
clean:
	rm -rf openshift-mirror-registry* image-archive.tar