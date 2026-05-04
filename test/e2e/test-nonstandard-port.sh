#!/usr/bin/env bash
# P2-4: Test installation with a non-standard port via --quayHostname.

set -euo pipefail
source "$(dirname "$0")/lib.sh"

MIRROR_REGISTRY="${1:-.}/mirror-registry"
HOSTNAME="${QUAY_HOSTNAME:-$(hostname -f)}"
CUSTOM_PORT="9443"
QUAY_ENDPOINT="${HOSTNAME}:${CUSTOM_PORT}"

log_info "=== Test: Non-Standard Port (${CUSTOM_PORT}) ==="

# Install with non-standard port
log_info "Installing with port ${CUSTOM_PORT}..."
${MIRROR_REGISTRY} install -v \
    --quayHostname "${QUAY_ENDPOINT}" \
    --initPassword password

wait_for_quay "${QUAY_ENDPOINT}"
assert_quay_healthy "${QUAY_ENDPOINT}"

# Verify the port is actually listening on the custom port
if ss -tlnp 2>/dev/null | grep -q ":${CUSTOM_PORT} "; then
    log_info "PASS: Port ${CUSTOM_PORT} is listening"
    ((PASS_COUNT++))
else
    log_error "FAIL: Port ${CUSTOM_PORT} is not listening"
    ((FAIL_COUNT++))
fi

# Verify default port 8443 is NOT listening
if ! ss -tlnp 2>/dev/null | grep -q ":8443 "; then
    log_info "PASS: Default port 8443 is not listening"
    ((PASS_COUNT++))
else
    log_warn "Port 8443 is also listening (may be expected if pod publishes both)"
fi

# Login and push/pull on custom port
assert_success "Podman login on port ${CUSTOM_PORT}" \
    podman login -u init -p password "${QUAY_ENDPOINT}" --tls-verify=false

verify_push_pull "${QUAY_ENDPOINT}" "init" "password" "--tls-verify=false"

# Cleanup
log_info "Uninstalling..."
${MIRROR_REGISTRY} uninstall --autoApprove -v

print_summary
