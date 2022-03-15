ARG RELEASE_VERSION=${RELEASE_VERSION}
ARG QUAY_IMAGE=${QUAY_IMAGE}
ARG EE_IMAGE=${EE_IMAGE}
ARG EE_BASE_IMAGE=${EE_BASE_IMAGE}
ARG EE_BUILDER_IMAGE=${EE_BUILDER_IMAGE}
ARG POSTGRES_IMAGE=${POSTGRES_IMAGE}
ARG REDIS_IMAGE=${REDIS_IMAGE}
ARG PAUSE_IMAGE=${PAUSE_IMAGE}

# Create Go CLI
FROM registry.redhat.io/ubi8:latest AS cli

# Need to duplicate these, otherwise they won't be available to the stage
ARG RELEASE_VERSION=${RELEASE_VERSION}
ARG QUAY_IMAGE=${QUAY_IMAGE}
ARG EE_IMAGE=${EE_IMAGE}
ARG POSTGRES_IMAGE=${POSTGRES_IMAGE}
ARG REDIS_IMAGE=${REDIS_IMAGE}
ARG PAUSE_IMAGE=${PAUSE_IMAGE}

ENV GOROOT=/usr/local/go
ENV PATH=$GOPATH/bin:$GOROOT/bin:$PATH 

# Get Go binary
RUN curl -o go1.16.4.linux-amd64.tar.gz https://dl.google.com/go/go1.16.4.linux-amd64.tar.gz
RUN tar -xzf go1.16.4.linux-amd64.tar.gz  &&\
    mv go /usr/local

COPY . /cli
WORKDIR /cli

# Create CLI
ENV RELEASE_VERSION=${RELEASE_VERSION}
ENV EE_IMAGE=${EE_IMAGE}
ENV QUAY_IMAGE=${QUAY_IMAGE}
ENV REDIS_IMAGE=${REDIS_IMAGE}
ENV POSTGRES_IMAGE=${POSTGRES_IMAGE}
ENV PAUSE_IMAGE=${PAUSE_IMAGE}

RUN go build -v \
	-ldflags "-X github.com/quay/mirror-registry/cmd.releaseVersion=${RELEASE_VERSION} -X github.com/quay/mirror-registry/cmd.eeImage=${EE_IMAGE} -X github.com/quay/mirror-registry/cmd.pauseImage=${PAUSE_IMAGE} -X github.com/quay/mirror-registry/cmd.quayImage=${QUAY_IMAGE} -X github.com/quay/mirror-registry/cmd.redisImage=${REDIS_IMAGE} -X github.com/quay/mirror-registry/cmd.postgresImage=${POSTGRES_IMAGE}" \
	-o mirror-registry

# Create Ansible Execution Environment
FROM $EE_BASE_IMAGE as galaxy
ARG ANSIBLE_GALAXY_CLI_COLLECTION_OPTS=
USER root

ADD ansible-runner/context/_build /build
WORKDIR /build

RUN ansible-galaxy role install -r requirements.yml --roles-path /usr/share/ansible/roles
RUN ansible-galaxy collection install $ANSIBLE_GALAXY_CLI_COLLECTION_OPTS -r requirements.yml --collections-path /usr/share/ansible/collections

FROM $EE_BUILDER_IMAGE as builder

COPY --from=galaxy /usr/share/ansible /usr/share/ansible

RUN ansible-builder introspect --sanitize --write-bindep=/tmp/src/bindep.txt --write-pip=/tmp/src/requirements.txt
RUN assemble

FROM $EE_BASE_IMAGE as ansible
USER root

COPY --from=galaxy /usr/share/ansible /usr/share/ansible

COPY --from=builder /output/ /output/
RUN /output/install-from-bindep && rm -rf /output/wheels
COPY ansible-runner/context/app /runner

# Pull in Quay dependencies
FROM $QUAY_IMAGE as quay
FROM $REDIS_IMAGE as redis
FROM $POSTGRES_IMAGE as postgres
FROM $PAUSE_IMAGE as pause

# Create mirror registry archive
FROM registry.redhat.io/ubi8:latest AS build

# Import and archive image dependencies
COPY --from=pause / /pause
RUN tar -cvf pause.tar -C /pause .

COPY --from=ansible / /ansible
RUN tar -cvf execution-environment.tar -C /ansible .

COPY --from=redis / /redis
RUN tar -cvf redis.tar -C /redis .

COPY --from=postgres / /postgres
RUN tar -cvf postgres.tar -C /postgres .

COPY --from=quay / /quay
RUN tar -cvf quay.tar -C /quay .

COPY --from=cli /cli/mirror-registry .

# Bundle quay, redis, postgres, and pause into a single archive
RUN tar -cvf image-archive.tar quay.tar redis.tar postgres.tar pause.tar

# Bundle mirror registry archive
RUN tar -czvf mirror-registry.tar.gz image-archive.tar execution-environment.tar mirror-registry

# Extract bundle to final release image
FROM registry.redhat.io/ubi8:latest AS release
COPY --from=build mirror-registry.tar.gz mirror-registry.tar.gz