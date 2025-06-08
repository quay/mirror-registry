# TREEISH is the hash of the current repository content including uncommitted
# changes, but not untracked files.
ifndef TREEISH
export TREEISH := $(shell tmpdir=$$(mktemp -d); export GIT_INDEX_FILE="$$tmpdir/index"; trap 'rm -rf $$tmpdir' EXIT; cp "$$(git rev-parse --git-dir)/index" "$$GIT_INDEX_FILE" && git add -u && git write-tree)
endif

# VERSION is the version of mirror-registry that should be installed. By
# default a build from the current repository is used.
VERSION := dev-$(TREEISH)

# OLD_VERSION is the version of mirror-registry that upgrade should be tested
# from.
OLD_VERSION := 1.2.9

# VAGRANT_UP_ARGS allows to provide additional arguments to `vagrant up`.
# Example:
#   make test-install VAGRANT_UP_ARGS='--provider=qemu'
ifeq ($(shell uname -s) $(shell uname -m),Darwin arm64)
VAGRANT_UP_ARGS := --provider=qemu
else
VAGRANT_UP_ARGS :=
endif

all: test-install test-sudo-install test-sudo-upgrade

# mirror-registry archive from the current repository.
mirror-registry-dev-$(TREEISH).tar.gz:
	# Building $@
	$(MAKE) -C .. build-offline-zip
	mv ../mirror-registry.tar.gz $@

# released mirror-registry archive.
mirror-registry-%.tar.gz:
	wget -O $@ https://developers.redhat.com/content-gateway/file/pub/openshift-v4/clients/mirror-registry/$*/mirror-registry.tar.gz

# inside-vagrant runs a command inside the Vagrant VM.
#
# The virtual machine will be starter before the command is run and stopped
# afterwards.
#
# Example:
#   $(call inside-vagrant,vagrant ssh -c 'uname -a')
define inside-vagrant
$(MAKE) start-vagrant && (set -x; $(1)); ret=$$?; $(MAKE) stop-vagrant && exit $$ret
endef

start-vagrant:
	vagrant up $(VAGRANT_UP_ARGS)

stop-vagrant:
ifneq ($(DEBUG), 1)
	vagrant destroy -f
endif

# vagrant-unpack uploads and unpacks the mirror-registry archive into the
# Vagrant VM.
vagrant-unpack: mirror-registry-$(VERSION).tar.gz
	vagrant upload mirror-registry-$(VERSION).tar.gz
	vagrant ssh -c "tar -vxf mirror-registry-$(VERSION).tar.gz"

vagrant-install: vagrant-unpack
	vagrant ssh -c "./mirror-registry install -v --initPassword password --quayHostname localhost"
	vagrant ssh -c "podman login -u init -p password localhost:8443 --tls-verify=false"

vagrant-sudo-install: vagrant-unpack
	vagrant ssh -c "sudo ./mirror-registry install -v --initPassword password --quayHostname localhost"
	vagrant ssh -c "podman login -u init -p password localhost:8443 --tls-verify=false"

vagrant-sudo-upgrade:
	$(MAKE) vagrant-sudo-install VERSION=$(OLD_VERSION)
	$(MAKE) vagrant-unpack
	vagrant ssh -c "sudo ./mirror-registry upgrade -v --quayHostname localhost"
	vagrant ssh -c "podman login -u init -p password localhost:8443 --tls-verify=false"

# test-install is an end-to-end test that installs mirror-registry.
# Version can be specified with VERSION.
# Use DEBUG=1 to prevent the Vagrant VM from being destroyed after the test.
#
# This target has an explicit dependency on the archive target so that the
# archive is built/downloaded before the virtual machine is started.
test-install: mirror-registry-$(VERSION).tar.gz
	$(call inside-vagrant,$(MAKE) vagrant-install)
	@echo "$@: OK"

test-sudo-install: mirror-registry-$(VERSION).tar.gz
	$(call inside-vagrant,$(MAKE) vagrant-sudo-install)
	@echo "$@: OK"

test-sudo-upgrade: mirror-registry-$(OLD_VERSION).tar.gz mirror-registry-$(VERSION).tar.gz
	$(call inside-vagrant,$(MAKE) vagrant-sudo-upgrade)
	@echo "$@: OK"
