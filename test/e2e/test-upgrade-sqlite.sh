#!/usr/bin/env bash
# P1-6: Test SQLite-to-SQLite upgrade path (recent release -> current build).
#
# Downloads the latest released version, installs it, pushes data, upgrades
# to the current PR build, and verifies data is preserved. This tests the
# common upgrade path (not the legacy postgres migration).

set -euo pipefail
source "$(dirname "$0")/lib.sh"

MIRROR_REGISTRY_DIR="${1:-.}"
HOSTNAME="${QUAY_HOSTNAME:-$(hostname -f)}"
QUAY_ENDPOINT="${HOSTNAME}:8443"

# Version to upgrade from — should be a recent SQLite-based release
OLD_VERSION="${OLD_VERSION:-2.0.2}"
OLD_TARBALL="mirror-registry-old.tar.gz"
OLD_DIR="mirror-registry-old"

log_info "=== Test: SQLite-to-SQLite Upgrade ==="
log_info "Upgrading from v${OLD_VERSION} to current build"

# Download old version
log_info "Downloading v${OLD_VERSION}..."
wget -q -O "${OLD_TARBALL}" \
    "https://github.com/quay/mirror-registry/releases/download/v${OLD_VERSION}/mirror-registry-offline.tar.gz" || {
    log_warn "Could not download v${OLD_VERSION}, skipping upgrade test"
    echo "SKIP: Old version not available for download"
    exit 0
}

mkdir -p "${OLD_DIR}"
tar -xzf "${OLD_TARBALL}" -C "${OLD_DIR}"

# Install old version
log_info "Installing v${OLD_VERSION}..."
./${OLD_DIR}/mirror-registry install -v \
    --quayHostname "${QUAY_ENDPOINT}" \
    --initPassword password

wait_for_quay "${QUAY_ENDPOINT}"

# Push test data
log_info "Pushing test image to old installation..."
podman login -u init -p password "${QUAY_ENDPOINT}" --tls-verify=false
podman pull docker.io/library/busybox 2>/dev/null || true
podman tag docker.io/library/busybox "${QUAY_ENDPOINT}/init/upgrade-test:v1"
podman push "${QUAY_ENDPOINT}/init/upgrade-test:v1" --tls-verify=false

# Upgrade to current build
log_info "Upgrading to current build..."
${MIRROR_REGISTRY_DIR}/mirror-registry upgrade -v \
    --quayHostname "${QUAY_ENDPOINT}"

wait_for_quay "${QUAY_ENDPOINT}"
assert_quay_healthy "${QUAY_ENDPOINT}"

# Verify data preserved — pull the image we pushed before upgrade
log_info "Verifying data survived upgrade..."
podman rmi "${QUAY_ENDPOINT}/init/upgrade-test:v1" 2>/dev/null || true
assert_success "Pull image pushed before upgrade" \
    podman pull "${QUAY_ENDPOINT}/init/upgrade-test:v1" --tls-verify=false

if podman images | grep -q "upgrade-test"; then
    log_info "PASS: Image data preserved through SQLite-to-SQLite upgrade"
    ((PASS_COUNT++))
else
    log_error "FAIL: Image data lost during upgrade"
    ((FAIL_COUNT++))
fi

# Verify login still works
assert_success "Login works after upgrade" \
    podman login -u init -p password "${QUAY_ENDPOINT}" --tls-verify=false

# Cleanup
log_info "Uninstalling..."
${MIRROR_REGISTRY_DIR}/mirror-registry uninstall --autoApprove -v
rm -rf "${OLD_DIR}" "${OLD_TARBALL}"

print_summary
