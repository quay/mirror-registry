#!/usr/bin/env bash
# P2-1: Test that install fails clearly when podman is not installed.

set -euo pipefail
source "$(dirname "$0")/lib.sh"

MIRROR_REGISTRY="${1:-.}/mirror-registry"
HOSTNAME="${QUAY_HOSTNAME:-$(hostname -f)}"
QUAY_ENDPOINT="${HOSTNAME}:8443"

log_info "=== Test: Error - Missing Podman ==="

# Remove podman
log_info "Removing podman..."
sudo yum -y remove podman 2>/dev/null || sudo dnf -y remove podman 2>/dev/null || true

# Verify podman is gone
if command -v podman &>/dev/null; then
    log_warn "podman still available, skipping test"
    # Reinstall for subsequent tests
    sudo yum -y install podman 2>/dev/null || sudo dnf -y install podman 2>/dev/null || true
    exit 0
fi

# Attempt install — should fail
log_info "Attempting install without podman..."
install_output=$( ${MIRROR_REGISTRY} install -v \
    --quayHostname "${QUAY_ENDPOINT}" \
    --initPassword password 2>&1 ) && install_rc=0 || install_rc=$?

if [[ ${install_rc} -ne 0 ]]; then
    log_info "PASS: Install failed as expected (exit code ${install_rc})"
    ((PASS_COUNT++))
else
    log_error "FAIL: Install succeeded without podman"
    ((FAIL_COUNT++))
fi

# Reinstall podman for subsequent tests
log_info "Reinstalling podman..."
sudo yum -y install podman 2>/dev/null || sudo dnf -y install podman 2>/dev/null || true

print_summary
