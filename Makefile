include .env

all:

build-golang-executable:
	sudo podman run --rm -v ${PWD}:/usr/src:Z -w /usr/src docker.io/golang:1.16 go build -v \
	-ldflags "-X github.com/quay/mirror-registry/cmd.eeImage=${EE_IMAGE} -X 'github.com/quay/mirror-registry/cmd.quayImage=${QUAY_IMAGE}' -X 'github.com/quay/mirror-registry/cmd.redisImage=${REDIS_IMAGE}' -X 'github.com/quay/mirror-registry/cmd.postgresImage=${POSTGRES_IMAGE}'" \
	-o mirror-registry;

build-online-zip: 
	sudo podman build \
		-t mirror-registry-online:${RELEASE_VERSION} \
		--build-arg QUAY_IMAGE=${QUAY_IMAGE} \
		--build-arg EE_IMAGE=${EE_IMAGE} \
		--build-arg EE_BASE_IMAGE=${EE_BASE_IMAGE} \
		--build-arg EE_BUILDER_IMAGE=${EE_BUILDER_IMAGE} \
		--build-arg POSTGRES_IMAGE=${POSTGRES_IMAGE} \
		--build-arg REDIS_IMAGE=${REDIS_IMAGE} \
		--build-arg PAUSE_IMAGE=${PAUSE_IMAGE} \
		--file Dockerfile.online . 
	sudo podman run --name mirror-registry-online-${RELEASE_VERSION} mirror-registry-online:${RELEASE_VERSION}
	sudo podman cp mirror-registry-online-${RELEASE_VERSION}:/mirror-registry.tar.gz .

build-offline-zip: 
	sudo podman build \
		-t mirror-registry-offline:${RELEASE_VERSION} \
		--build-arg QUAY_IMAGE=${QUAY_IMAGE} \
		--build-arg EE_IMAGE=${EE_IMAGE} \
		--build-arg EE_BASE_IMAGE=${EE_BASE_IMAGE} \
		--build-arg EE_BUILDER_IMAGE=${EE_BUILDER_IMAGE} \
		--build-arg POSTGRES_IMAGE=${POSTGRES_IMAGE} \
		--build-arg REDIS_IMAGE=${REDIS_IMAGE} \
		--build-arg PAUSE_IMAGE=${PAUSE_IMAGE} \
		--file Dockerfile .
	sudo podman run --name mirror-registry-offline-${RELEASE_VERSION} mirror-registry-offline:${RELEASE_VERSION}
	sudo podman cp mirror-registry-offline-${RELEASE_VERSION}:/mirror-registry.tar.gz .

clean:
	rm -rf mirror-registry* image-archive.tar
