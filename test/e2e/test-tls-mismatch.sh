#!/usr/bin/env bash
# P1-7: Test TLS certificate hostname mismatch and --sslCheckSkip.
#
# Generates a certificate for a wrong hostname, verifies install rejects it,
# then re-tests with --sslCheckSkip to confirm the bypass works.

set -euo pipefail
source "$(dirname "$0")/lib.sh"

MIRROR_REGISTRY="${1:-.}/mirror-registry"
HOSTNAME="${QUAY_HOSTNAME:-$(hostname -f)}"
QUAY_ENDPOINT="${HOSTNAME}:8443"
CERT_DIR="/tmp/test-tls-mismatch"
WRONG_HOSTNAME="wrong-host.example.com"

log_info "=== Test: TLS Certificate Hostname Mismatch ==="
log_info "Actual hostname: ${HOSTNAME}"
log_info "Certificate hostname: ${WRONG_HOSTNAME}"

# Generate cert for the WRONG hostname
generate_test_certs "${CERT_DIR}" "${WRONG_HOSTNAME}"

# Attempt install with mismatched cert — should fail
log_info "Installing with mismatched certificate (should fail)..."
if ${MIRROR_REGISTRY} install -v \
    --quayHostname "${QUAY_ENDPOINT}" \
    --initPassword password \
    --sslCert "${CERT_DIR}/server.cert" \
    --sslKey "${CERT_DIR}/server.key" 2>&1; then
    log_error "FAIL: Install should have rejected mismatched certificate"
    ((FAIL_COUNT++))
    # Clean up if it somehow installed
    ${MIRROR_REGISTRY} uninstall --autoApprove -v 2>/dev/null || true
else
    log_info "PASS: Install correctly rejected mismatched certificate"
    ((PASS_COUNT++))
fi

# Now install with --sslCheckSkip — should succeed
log_info "Installing with --sslCheckSkip (should succeed)..."
${MIRROR_REGISTRY} install -v \
    --quayHostname "${QUAY_ENDPOINT}" \
    --initPassword password \
    --sslCert "${CERT_DIR}/server.cert" \
    --sslKey "${CERT_DIR}/server.key" \
    --sslCheckSkip

wait_for_quay "${QUAY_ENDPOINT}"
assert_quay_healthy "${QUAY_ENDPOINT}"
log_info "PASS: Install succeeded with --sslCheckSkip"
((PASS_COUNT++))

# Verify we can still login (using tls-verify=false since cert is for wrong host)
assert_success "Login works with --sslCheckSkip install" \
    podman login -u init -p password "${QUAY_ENDPOINT}" --tls-verify=false

# Cleanup
log_info "Uninstalling..."
${MIRROR_REGISTRY} uninstall --autoApprove -v
rm -rf "${CERT_DIR}"

print_summary
