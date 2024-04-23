ARG RELEASE_VERSION=${RELEASE_VERSION}
ARG QUAY_IMAGE=${QUAY_IMAGE}
ARG EE_IMAGE=${EE_IMAGE}
ARG EE_BASE_IMAGE=${EE_BASE_IMAGE}
ARG EE_BUILDER_IMAGE=${EE_BUILDER_IMAGE}
ARG REDIS_IMAGE=${REDIS_IMAGE}
ARG PAUSE_IMAGE=${PAUSE_IMAGE}

# Create Go CLI
FROM registry.access.redhat.com/ubi8:latest AS cli

# Need to duplicate these, otherwise they won't be available to the stage
ARG RELEASE_VERSION=${RELEASE_VERSION}
ARG QUAY_IMAGE=${QUAY_IMAGE}
ARG EE_IMAGE=${EE_IMAGE}
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
ENV PAUSE_IMAGE=${PAUSE_IMAGE}

RUN go build -v \
<<<<<<< HEAD
	-ldflags "-X github.com/quay/mirror-registry/cmd.releaseVersion=${RELEASE_VERSION} -X github.com/quay/mirror-registry/cmd.eeImage=${EE_IMAGE} -X github.com/quay/mirror-registry/cmd.pauseImage=${PAUSE_IMAGE} -X github.com/quay/mirror-registry/cmd.quayImage=${QUAY_IMAGE} -X github.com/quay/mirror-registry/cmd.redisImage=${REDIS_IMAGE}" \
=======
	-ldflags "-X github.com/quay/mirror-registry/cmd.releaseVersion=${RELEASE_VERSION} -X github.com/quay/mirror-registry/cmd.eeImage=${EE_IMAGE} -X github.com/quay/mirror-registry/cmd.pauseImage=${PAUSE_IMAGE} -X github.com/quay/mirror-registry/cmd.quayImage=${QUAY_IMAGE} -X github.com/quay/mirror-registry/cmd.redisImage=${REDIS_IMAGE} -X github.com/quay/mirror-registry/cmd.postgresImage=${POSTGRES_IMAGE} -X github.com/quay/mirror-registry/cmd.sqliteImage=${DB_TO_SQLITE_IMAGE}" \
>>>>>>> 7261fe9 (Add support for sqlite storage in installer binary (PROJQUAY-6286))
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
FROM $PAUSE_IMAGE as pause

# Install db sqlite migration cli
FROM registry.access.redhat.com/ubi8-minimal AS db-cli

# Install Python 3 and pip
RUN microdnf update -y && \
    microdnf install python3 python3-pip -y && \
    microdnf install gcc-c++ python3-devel -y && \
    microdnf clean all

COPY requirements.txt .
RUN pip3 install --no-cache-dir -r requirements.txt

# Create mirror registry archive
FROM registry.access.redhat.com/ubi8:latest AS build

# Import and archive image dependencies
COPY --from=pause / /pause
RUN tar -cvf pause.tar -C /pause .

COPY --from=ansible / /ansible
RUN tar -cvf execution-environment.tar -C /ansible .

COPY --from=redis / /redis
RUN tar -cvf redis.tar -C /redis .

COPY --from=quay / /quay
RUN tar -cvf quay.tar -C /quay .

COPY --from=cli /cli/mirror-registry .

COPY --from=db-cli / /db-cli
RUN tar -cvf db-cli.tar -C /db-cli .

# Bundle quay, redis and pause into a single archive
RUN tar -cvf image-archive.tar quay.tar redis.tar pause.tar

# Bundle mirror registry archive
RUN tar -czvf mirror-registry.tar.gz image-archive.tar execution-environment.tar mirror-registry db-cli.tar

# Extract bundle to final release image
FROM registry.access.redhat.com/ubi8:latest AS release
COPY --from=build mirror-registry.tar.gz mirror-registry.tar.gz
