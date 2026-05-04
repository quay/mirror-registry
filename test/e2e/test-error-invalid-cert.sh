#!/usr/bin/env bash
# P2-3: Test that install fails with a clear error for malformed certificates.

set -euo pipefail
source "$(dirname "$0")/lib.sh"

MIRROR_REGISTRY="${1:-.}/mirror-registry"
HOSTNAME="${QUAY_HOSTNAME:-$(hostname -f)}"
QUAY_ENDPOINT="${HOSTNAME}:8443"
CERT_DIR="/tmp/test-invalid-certs"

log_info "=== Test: Error - Invalid Certificate ==="

mkdir -p "${CERT_DIR}"

# Test 1: Malformed cert file (not PEM)
log_info "Test 1: Malformed certificate file..."
echo "this is not a certificate" > "${CERT_DIR}/bad.cert"
echo "this is not a key" > "${CERT_DIR}/bad.key"

install_output=$( ${MIRROR_REGISTRY} install -v \
    --quayHostname "${QUAY_ENDPOINT}" \
    --initPassword password \
    --sslCert "${CERT_DIR}/bad.cert" \
    --sslKey "${CERT_DIR}/bad.key" 2>&1 ) && install_rc=0 || install_rc=$?

if [[ ${install_rc} -ne 0 ]]; then
    log_info "PASS: Install rejected malformed certificate"
    ((PASS_COUNT++))
else
    log_error "FAIL: Install accepted malformed certificate"
    ((FAIL_COUNT++))
    ${MIRROR_REGISTRY} uninstall --autoApprove -v 2>/dev/null || true
fi

# Test 2: Nonexistent cert file
log_info "Test 2: Nonexistent certificate file..."
install_output=$( ${MIRROR_REGISTRY} install -v \
    --quayHostname "${QUAY_ENDPOINT}" \
    --initPassword password \
    --sslCert "/nonexistent/cert.pem" \
    --sslKey "/nonexistent/key.pem" 2>&1 ) && install_rc=0 || install_rc=$?

if [[ ${install_rc} -ne 0 ]]; then
    log_info "PASS: Install rejected nonexistent certificate files"
    ((PASS_COUNT++))
else
    log_error "FAIL: Install accepted nonexistent certificate files"
    ((FAIL_COUNT++))
    ${MIRROR_REGISTRY} uninstall --autoApprove -v 2>/dev/null || true
fi

# Cleanup
rm -rf "${CERT_DIR}"

print_summary
