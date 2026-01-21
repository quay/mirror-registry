TEST_DIR ?= .

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

test-all: test-install test-sudo-install test-sudo-upgrade

# mirror-registry archive from the current repository.
$(TEST_DIR)/mirror-registry-dev-$(TREEISH).tar.gz:
	# Building $@
	$(MAKE) build-offline-zip
	mv mirror-registry.tar.gz $@

# released mirror-registry archive.
$(TEST_DIR)/mirror-registry-%.tar.gz:
	wget -O $@ https://developers.redhat.com/content-gateway/file/pub/openshift-v4/clients/mirror-registry/$*/mirror-registry.tar.gz

# inside-vagrant runs a command inside the Vagrant VM.
#
# The virtual machine will be started before the command is run and stopped
# afterwards.
#
# Example:
#   $(call inside-vagrant,vagrant ssh -c 'uname -a')
define inside-vagrant
$(MAKE) start-vagrant && (set -x; $(1)); ret=$$?; $(MAKE) stop-vagrant && exit $$ret
endef

start-vagrant:
	cd $(TEST_DIR) && vagrant up

stop-vagrant:
ifneq ($(DEBUG), 1)
	cd $(TEST_DIR) && vagrant destroy -f
endif

vagrant-ssh:
	cd $(TEST_DIR) && vagrant ssh

# vagrant-unpack uploads and unpacks the mirror-registry archive into the
# Vagrant VM.
vagrant-unpack: $(TEST_DIR)/mirror-registry-$(VERSION).tar.gz
	cd $(TEST_DIR) && vagrant upload mirror-registry-$(VERSION).tar.gz
	cd $(TEST_DIR) && vagrant ssh -c "tar -vxf mirror-registry-$(VERSION).tar.gz"

vagrant-install: vagrant-unpack
	cd $(TEST_DIR) && vagrant ssh -c "./mirror-registry install -v --initPassword password --quayHostname localhost"
	cd $(TEST_DIR) && vagrant ssh -c "podman login -u init -p password localhost:8443 --tls-verify=false"

vagrant-sudo-install: vagrant-unpack
	cd $(TEST_DIR) && vagrant ssh -c "sudo ./mirror-registry install -v --initPassword password --quayHostname localhost"
	cd $(TEST_DIR) && vagrant ssh -c "podman login -u init -p password localhost:8443 --tls-verify=false"

vagrant-sudo-upgrade:
	$(MAKE) vagrant-sudo-install VERSION=$(OLD_VERSION)
	$(MAKE) vagrant-unpack
	cd $(TEST_DIR) && vagrant ssh -c "sudo ./mirror-registry upgrade -v --quayHostname localhost"
	cd $(TEST_DIR) && vagrant ssh -c "podman login -u init -p password localhost:8443 --tls-verify=false"

# test-install is an end-to-end test that installs mirror-registry.
# Version can be specified with VERSION.
# Use DEBUG=1 to prevent the Vagrant VM from being destroyed after the test.
#
# This target has an explicit dependency on the archive target so that the
# archive is built/downloaded before the virtual machine is started.
test-install: $(TEST_DIR)/mirror-registry-$(VERSION).tar.gz
	$(call inside-vagrant,$(MAKE) vagrant-install)
	@echo "$@: OK"

test-sudo-install: $(TEST_DIR)/mirror-registry-$(VERSION).tar.gz
	$(call inside-vagrant,$(MAKE) vagrant-sudo-install)
	@echo "$@: OK"

test-sudo-upgrade: $(TEST_DIR)/mirror-registry-$(OLD_VERSION).tar.gz $(TEST_DIR)/mirror-registry-$(VERSION).tar.gz
	$(call inside-vagrant,$(MAKE) vagrant-sudo-upgrade)
	@echo "$@: OK"
