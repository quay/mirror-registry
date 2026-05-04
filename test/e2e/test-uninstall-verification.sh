#!/usr/bin/env bash
# P0-9: Comprehensive uninstall verification.
#
# Installs mirror-registry, pushes an image, then uninstalls and verifies
# that all artifacts are completely cleaned up.

set -euo pipefail
source "$(dirname "$0")/lib.sh"

MIRROR_REGISTRY="${1:-.}/mirror-registry"
HOSTNAME="${QUAY_HOSTNAME:-$(hostname -f)}"
QUAY_ENDPOINT="${HOSTNAME}:8443"

log_info "=== Test: Uninstall Completeness Verification ==="
log_info "Hostname: ${HOSTNAME}"

# Install
log_info "Installing..."
${MIRROR_REGISTRY} install -v \
    --quayHostname "${QUAY_ENDPOINT}" \
    --initPassword password

wait_for_quay "${QUAY_ENDPOINT}"

# Push a test image to ensure data exists
verify_push_pull "${QUAY_ENDPOINT}" "init" "password" "--tls-verify=false"

# Uninstall
log_info "Uninstalling with --autoApprove..."
${MIRROR_REGISTRY} uninstall --autoApprove -v

# Comprehensive cleanup checks
assert_clean_uninstall

# Verify no quay-related podman pods exist
pods=$(podman pod ls --format "{{.Name}}" 2>/dev/null | grep -i quay || true)
if [[ -z "${pods}" ]]; then
    log_info "PASS: No quay podman pods exist"
    ((PASS_COUNT++))
else
    log_error "FAIL: Quay pods still exist: ${pods}"
    ((FAIL_COUNT++))
fi

# Verify port 8443 is no longer in use
if ! ss -tlnp 2>/dev/null | grep -q ":8443 "; then
    log_info "PASS: Port 8443 is free"
    ((PASS_COUNT++))
else
    log_error "FAIL: Port 8443 is still in use"
    ((FAIL_COUNT++))
fi

# Verify the health endpoint is no longer accessible
if ! curl -sk --connect-timeout 5 "https://${QUAY_ENDPOINT}/health/instance" &>/dev/null; then
    log_info "PASS: Health endpoint is no longer accessible"
    ((PASS_COUNT++))
else
    log_error "FAIL: Health endpoint still accessible after uninstall"
    ((FAIL_COUNT++))
fi

print_summary
