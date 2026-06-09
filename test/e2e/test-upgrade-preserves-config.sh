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
${MIRROR_REGISTRY} install -v \
    --quayHostname "${QUAY_ENDPOINT}" \
    --initPassword password

wait_for_quay "${QUAY_ENDPOINT}"

# Record pre-upgrade state
pre_hostname=$(grep SERVER_HOSTNAME ~/quay-install/quay-config/config.yaml | awk '{print $2}')
log_info "Pre-upgrade SERVER_HOSTNAME: ${pre_hostname}"

# Push test image
podman login -u init -p password "${QUAY_ENDPOINT}" --tls-verify=false
podman pull docker.io/library/busybox 2>/dev/null || true
podman tag docker.io/library/busybox "${QUAY_ENDPOINT}/init/config-test:v1"
podman push "${QUAY_ENDPOINT}/init/config-test:v1" --tls-verify=false

# Upgrade WITHOUT passing --quayHostname (the bug scenario)
log_info "Upgrading without --quayHostname flag..."
${MIRROR_REGISTRY} upgrade -v

# Verify port preserved
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
post_hostname=$(grep SERVER_HOSTNAME ~/quay-install/quay-config/config.yaml | awk '{print $2}')
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

# Cleanup
${MIRROR_REGISTRY} uninstall --autoApprove -v

print_summary
