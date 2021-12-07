ARG EE_BASE_IMAGE=registry.redhat.io/ansible-automation-platform-20-early-access/ee-minimal-rhel8
ARG EE_BUILDER_IMAGE=registry.redhat.io/ansible-automation-platform-20-early-access/ansible-builder-rhel8:2.0.0-15

# Create Go CLI
FROM registry.redhat.io/ubi8:latest AS cli

ENV GOROOT=/usr/local/go
ENV PATH=$GOPATH/bin:$GOROOT/bin:$PATH 

# Get Go binary
RUN curl -o go1.16.4.linux-amd64.tar.gz https://dl.google.com/go/go1.16.4.linux-amd64.tar.gz
RUN tar -xzf go1.16.4.linux-amd64.tar.gz  &&\
    mv go /usr/local

COPY . /cli
WORKDIR /cli

# Create CLI
ENV EE_IMAGE=quay.io/quay/openshift-mirror-registry-ee:latest
ENV QUAY_IMAGE=registry.redhat.io/quay/quay-rhel8:v3.6.1
ENV REDIS_IMAGE=registry.redhat.io/rhel8/redis-6:1-25
ENV POSTGRES_IMAGE=registry.redhat.io/rhel8/postgresql-10:1-161

RUN go build -v \
	-ldflags "-X github.com/quay/openshift-mirror-registry/cmd.eeImage=${EE_IMAGE} -X github.com/quay/openshift-mirror-registry/cmd.quayImage=${QUAY_IMAGE} -X github.com/quay/openshift-mirror-registry/cmd.redisImage=${REDIS_IMAGE} -X github.com/quay/openshift-mirror-registry/cmd.postgresImage=${POSTGRES_IMAGE}" \
	-o openshift-mirror-registry

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

ADD ansible-runner/context/_build/requirements.txt requirements.txt
RUN ansible-builder introspect --sanitize --user-pip=requirements.txt --write-bindep=/tmp/src/bindep.txt --write-pip=/tmp/src/requirements.txt
RUN assemble

FROM $EE_BASE_IMAGE as ansible
USER root

COPY --from=galaxy /usr/share/ansible /usr/share/ansible

COPY --from=builder /output/ /output/
RUN /output/install-from-bindep && rm -rf /output/wheels
COPY ansible-runner/context/app /runner

# Pull in Quay dependencies
FROM registry.redhat.io/quay/quay-rhel8:v3.6.1 as quay
FROM registry.redhat.io/rhel8/redis-6:1-25 as redis
FROM registry.redhat.io/rhel8/postgresql-10:1-161 as postgres

# Create OMR archive
FROM registry.redhat.io/ubi8:latest AS build

# Import and archive image dependencies
COPY --from=ansible / /ansible
RUN tar -cvf execution-environment.tar -C /ansible .

COPY --from=redis / /redis
RUN tar -cvf redis.tar -C /redis .

COPY --from=postgres / /postgres
RUN tar -cvf postgres.tar -C /postgres .

COPY --from=quay / /quay
RUN tar -cvf quay.tar -C /quay .

COPY --from=cli /cli/openshift-mirror-registry .

# Bundle quay, redis, and postgres into a single archive
RUN tar -cvf image-archive.tar quay.tar redis.tar postgres.tar

# Bundle OMR archive
RUN tar -czvf openshift-mirror-registry.tar.gz image-archive.tar execution-environment.tar openshift-mirror-registry

# Extract bundle to final release image
FROM registry.redhat.io/ubi8:latest AS release
COPY --from=build openshift-mirror-registry.tar.gz openshift-mirror-registry.tar.gz