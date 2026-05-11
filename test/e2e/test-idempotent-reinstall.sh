#!/usr/bin/env bash
# P1-5: Test idempotent reinstall (install over existing without uninstall).
#
# Installs mirror-registry, pushes data, then installs again without
# uninstalling first. Verifies the second install either succeeds
# gracefully or fails with a clear error.

set -euo pipefail
source "$(dirname "$0")/lib.sh"

MIRROR_REGISTRY="${1:-.}/mirror-registry"
HOSTNAME="${QUAY_HOSTNAME:-$(hostname -f)}"
QUAY_ENDPOINT="${HOSTNAME}:8443"

log_info "=== Test: Idempotent Reinstall ==="

# First install
log_info "First install..."
${MIRROR_REGISTRY} install -v \
    --quayHostname "${QUAY_ENDPOINT}" \
    --initPassword password

wait_for_quay "${QUAY_ENDPOINT}"

# Push test data
log_info "Pushing test image..."
podman login -u init -p password "${QUAY_ENDPOINT}" --tls-verify=false
podman pull docker.io/library/busybox 2>/dev/null || true
podman tag docker.io/library/busybox "${QUAY_ENDPOINT}/init/reinstall-test:v1"
podman push "${QUAY_ENDPOINT}/init/reinstall-test:v1" --tls-verify=false

# Second install over existing (without uninstall)
log_info "Second install (over existing)..."
reinstall_output=$( ${MIRROR_REGISTRY} install -v \
    --quayHostname "${QUAY_ENDPOINT}" \
    --initPassword password 2>&1 ) && reinstall_rc=0 || reinstall_rc=$?

if [[ ${reinstall_rc} -eq 0 ]]; then
    log_info "PASS: Reinstall succeeded (idempotent behavior)"
    ((PASS_COUNT++))

    wait_for_quay "${QUAY_ENDPOINT}"
    assert_quay_healthy "${QUAY_ENDPOINT}"
else
    # If reinstall fails, it should at least produce a clear error
    log_warn "Reinstall failed with exit code ${reinstall_rc}"
    if echo "${reinstall_output}" | grep -qi "already\|exist\|running\|conflict"; then
        log_info "PASS: Reinstall failed with descriptive error message"
        ((PASS_COUNT++))
    else
        log_error "FAIL: Reinstall failed without a clear error message"
        log_error "  Output: ${reinstall_output}"
        ((FAIL_COUNT++))
    fi
fi

# Cleanup
log_info "Uninstalling..."
${MIRROR_REGISTRY} uninstall --autoApprove -v

print_summary
