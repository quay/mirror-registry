#!/usr/bin/env bash
# P1-2/P1-3: Test installation with custom quayRoot and filesystem storage paths.
#
# Installs with --quayRoot, --quayStorage, --sqliteStorage set to filesystem
# paths instead of default named volumes. Verifies directories are created,
# push/pull works, and uninstall cleans everything up.

set -euo pipefail
source "$(dirname "$0")/lib.sh"

MIRROR_REGISTRY="${1:-.}/mirror-registry"
HOSTNAME="${QUAY_HOSTNAME:-$(hostname -f)}"
QUAY_ENDPOINT="${HOSTNAME}:8443"

CUSTOM_ROOT="/opt/quay-data"
CUSTOM_STORAGE="/data/quay-blobs"
CUSTOM_SQLITE="/data/quay-db"

log_info "=== Test: Custom Paths Install ==="
log_info "quayRoot: ${CUSTOM_ROOT}"
log_info "quayStorage: ${CUSTOM_STORAGE}"
log_info "sqliteStorage: ${CUSTOM_SQLITE}"

# Create parent directories
sudo mkdir -p /opt /data
sudo chmod 777 /opt /data

# Install with custom paths
log_info "Installing with custom paths..."
${MIRROR_REGISTRY} install -v \
    --quayHostname "${QUAY_ENDPOINT}" \
    --initPassword password \
    --quayRoot "${CUSTOM_ROOT}" \
    --quayStorage "${CUSTOM_STORAGE}" \
    --sqliteStorage "${CUSTOM_SQLITE}"

wait_for_quay "${QUAY_ENDPOINT}"
assert_quay_healthy "${QUAY_ENDPOINT}"

# Verify custom directories exist
assert_file_exists "quayRoot directory created" "${CUSTOM_ROOT}"
assert_file_exists "quayStorage directory created" "${CUSTOM_STORAGE}"
assert_file_exists "sqliteStorage directory created" "${CUSTOM_SQLITE}"

# Verify config files are in custom root
assert_file_exists "config.yaml in custom quayRoot" "${CUSTOM_ROOT}/quay-config/config.yaml"

# Push/pull test
verify_push_pull "${QUAY_ENDPOINT}" "init" "password" "--tls-verify=false"

# Uninstall with custom paths
log_info "Uninstalling with custom paths..."
${MIRROR_REGISTRY} uninstall --autoApprove -v \
    --quayRoot "${CUSTOM_ROOT}" \
    --quayStorage "${CUSTOM_STORAGE}" \
    --sqliteStorage "${CUSTOM_SQLITE}"

# Verify cleanup
assert_file_not_exists "quayRoot removed after uninstall" "${CUSTOM_ROOT}"

# Verify no services remain
local_containers=$(podman ps -q -f name=quay 2>/dev/null || true)
if [[ -z "${local_containers}" ]]; then
    log_info "PASS: No quay containers running after uninstall"
    ((PASS_COUNT++))
else
    log_error "FAIL: Quay containers still running"
    ((FAIL_COUNT++))
fi

print_summary
