include .env

all:

build-golang-executable:
	sudo podman run --rm -v ${PWD}:/usr/src:Z -w /usr/src docker.io/golang:1.16 go build -v \
	-ldflags "-X github.com/quay/openshift-mirror-registry/cmd.eeImage=${EE_IMAGE} -X 'github.com/quay/openshift-mirror-registry/cmd.quayImage=${QUAY_IMAGE}' -X 'github.com/quay/openshift-mirror-registry/cmd.redisImage=${REDIS_IMAGE}' -X 'github.com/quay/openshift-mirror-registry/cmd.postgresImage=${POSTGRES_IMAGE}'" \
	-o openshift-mirror-registry;

build-online-zip: 
	sudo podman build -t omr-online:${RELEASE_VERSION} --file Dockerfile.online . 
	sudo podman run --name omr-online-${RELEASE_VERSION} omr-online:${RELEASE_VERSION}
	sudo podman cp omr-online-${RELEASE_VERSION}:/openshift-mirror-registry.tar.gz .

build-offline-zip: 
	sudo podman build -t omr-offline:${RELEASE_VERSION} --file Dockerfile .
	sudo podman run --name omr-offline-${RELEASE_VERSION} omr-offline:${RELEASE_VERSION}
	sudo podman cp omr-offline-${RELEASE_VERSION}:/openshift-mirror-registry.tar.gz .

release:
	git add .
	git commit -m "release: Release Version ${RELEASE_VERSION}"
	git push

clean:
	rm -rf openshift-mirror-registry* image-archive.tar