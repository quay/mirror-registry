#!/usr/bin/env bash
# Test that upgrade preserves existing hostname/port and storage paths
# when flags are not explicitly re-passed.

set -euo pipefail
source "$(dirname "$0")/lib.sh"

MIRROR_REGISTRY="${1:-.}/mirror-registry"
HOSTNAME="${QUAY_HOSTNAME:-$(hostname -f)}"
CUSTOM_PORT="9443"
QUAY_ENDPOINT="${HOSTNAME}:${CUSTOM_PORT}"

log_info "=== Test: Upgrade Preserves Existing Config ==="

# Install with custom port
log_info "Installing with custom port ${CUSTOM_PORT}..."
"${MIRROR_REGISTRY}" install -v \
    --quayHostname "${QUAY_ENDPOINT}" \
    --initPassword password

wait_for_quay "${QUAY_ENDPOINT}"

# Record pre-upgrade state
pre_hostname=$(grep '^SERVER_HOSTNAME' ~/quay-install/quay-config/config.yaml | awk '{print $2}')
pre_storage=$(systemctl --user cat quay-app.service 2>/dev/null | grep -oP '\S+(?=:/datastorage)' | head -1)
log_info "Pre-upgrade SERVER_HOSTNAME: ${pre_hostname}"
log_info "Pre-upgrade datastorage mount: ${pre_storage}"

# Push test image
podman login -u init -p password "${QUAY_ENDPOINT}" --tls-verify=false
podman pull docker.io/library/busybox 2>/dev/null || true
podman tag docker.io/library/busybox "${QUAY_ENDPOINT}/init/config-test:v1"
podman push "${QUAY_ENDPOINT}/init/config-test:v1" --tls-verify=false

# --- Test 1: Upgrade without flags preserves hostname ---
log_info "Upgrading without --quayHostname flag..."
"${MIRROR_REGISTRY}" upgrade -v

wait_for_quay "${QUAY_ENDPOINT}"
assert_quay_healthy "${QUAY_ENDPOINT}"

if ss -tlnp 2>/dev/null | grep -q ":${CUSTOM_PORT} "; then
    log_info "PASS: Custom port ${CUSTOM_PORT} preserved after upgrade"
    (( ++PASS_COUNT ))
else
    log_error "FAIL: Custom port ${CUSTOM_PORT} not listening after upgrade"
    (( ++FAIL_COUNT ))
fi

# Verify SERVER_HOSTNAME preserved in config.yaml
post_hostname=$(grep '^SERVER_HOSTNAME' ~/quay-install/quay-config/config.yaml | awk '{print $2}')
if [[ "${pre_hostname}" == "${post_hostname}" ]]; then
    log_info "PASS: SERVER_HOSTNAME preserved (${post_hostname})"
    (( ++PASS_COUNT ))
else
    log_error "FAIL: SERVER_HOSTNAME changed from ${pre_hostname} to ${post_hostname}"
    (( ++FAIL_COUNT ))
fi

# Verify data survived
podman rmi "${QUAY_ENDPOINT}/init/config-test:v1" 2>/dev/null || true
assert_success "Pull image after upgrade" \
    podman pull "${QUAY_ENDPOINT}/init/config-test:v1" --tls-verify=false

# Verify storage source path preserved in quay-app.service
post_storage=$(systemctl --user cat quay-app.service 2>/dev/null | grep -oP '\S+(?=:/datastorage)' | head -1)
if [[ -n "${pre_storage}" && "${pre_storage}" == "${post_storage}" ]]; then
    log_info "PASS: quay-storage mount preserved (${post_storage})"
    (( ++PASS_COUNT ))
else
    log_error "FAIL: quay-storage mount changed from ${pre_storage} to ${post_storage}"
    (( ++FAIL_COUNT ))
fi

# --- Test 2: Explicit --quayHostname override still works ---
log_info "Upgrading with explicit --quayHostname ${HOSTNAME}:7443..."
"${MIRROR_REGISTRY}" upgrade --quayHostname "${HOSTNAME}:7443" -v

wait_for_quay "${HOSTNAME}:7443"

if ss -tlnp 2>/dev/null | grep -q ":7443 "; then
    log_info "PASS: Explicit --quayHostname override to port 7443 works"
    (( ++PASS_COUNT ))
else
    log_error "FAIL: Port 7443 not listening after explicit override"
    (( ++FAIL_COUNT ))
fi

# Cleanup test 1+2
"${MIRROR_REGISTRY}" uninstall --autoApprove -v

# --- Test 3: Custom quayRoot preserved on upgrade ---
CUSTOM_ROOT="/tmp/omr-test-root-$$"
mkdir -p "${CUSTOM_ROOT}"
log_info "Installing with custom quayRoot ${CUSTOM_ROOT}..."
"${MIRROR_REGISTRY}" install -v \
    --quayHostname "${HOSTNAME}:8443" \
    --initPassword password \
    --quayRoot "${CUSTOM_ROOT}"

wait_for_quay "${HOSTNAME}:8443"

# Verify config.yaml is at the custom path
if [[ -f "${CUSTOM_ROOT}/quay-config/config.yaml" ]]; then
    log_info "PASS: config.yaml exists at custom quayRoot"
    (( ++PASS_COUNT ))
else
    log_error "FAIL: config.yaml not found at ${CUSTOM_ROOT}/quay-config/config.yaml"
    (( ++FAIL_COUNT ))
fi

# Upgrade without --quayRoot
log_info "Upgrading without --quayRoot flag..."
"${MIRROR_REGISTRY}" upgrade -v

wait_for_quay "${HOSTNAME}:8443"
assert_quay_healthy "${HOSTNAME}:8443"

# Verify config.yaml still at custom path (not default ~/quay-install)
if [[ -f "${CUSTOM_ROOT}/quay-config/config.yaml" ]]; then
    log_info "PASS: Custom quayRoot preserved after upgrade (${CUSTOM_ROOT})"
    (( ++PASS_COUNT ))
else
    log_error "FAIL: Custom quayRoot lost — config.yaml not at ${CUSTOM_ROOT}"
    (( ++FAIL_COUNT ))
fi

# Cleanup test 3
"${MIRROR_REGISTRY}" uninstall --autoApprove -v --quayRoot "${CUSTOM_ROOT}"
rm -rf "${CUSTOM_ROOT}"

print_summary
