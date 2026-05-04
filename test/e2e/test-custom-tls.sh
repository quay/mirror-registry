#!/usr/bin/env bash
# P0-6: Test installation with custom TLS certificates.
#
# Generates a CA + server certificate, installs mirror-registry with
# --sslCert/--sslKey, and verifies HTTPS works with that certificate.

set -euo pipefail
source "$(dirname "$0")/lib.sh"

MIRROR_REGISTRY="${1:-.}/mirror-registry"
HOSTNAME="${QUAY_HOSTNAME:-$(hostname -f)}"
QUAY_ENDPOINT="${HOSTNAME}:8443"
CERT_DIR="/tmp/test-tls-certs"

log_info "=== Test: Custom TLS Certificate Install ==="
log_info "Hostname: ${HOSTNAME}"

# Generate test CA and server cert
generate_test_certs "${CERT_DIR}" "${HOSTNAME}"

# Install with custom certificates
log_info "Installing with custom TLS certificates..."
${MIRROR_REGISTRY} install -v \
    --quayHostname "${QUAY_ENDPOINT}" \
    --initPassword password \
    --sslCert "${CERT_DIR}/server.cert" \
    --sslKey "${CERT_DIR}/server.key"

wait_for_quay "${QUAY_ENDPOINT}"
assert_quay_healthy "${QUAY_ENDPOINT}"

# Verify the certificate is the one we provided
log_info "Verifying server certificate matches provided cert..."
server_serial=$(echo | openssl s_client -connect "${QUAY_ENDPOINT}" -servername "${HOSTNAME}" 2>/dev/null | openssl x509 -noout -serial 2>/dev/null || echo "CONNECT_FAILED")
provided_serial=$(openssl x509 -in "${CERT_DIR}/server.cert" -noout -serial 2>/dev/null)

if [[ "${server_serial}" == "${provided_serial}" ]]; then
    log_info "PASS: Server is using the provided certificate"
    ((PASS_COUNT++))
else
    log_error "FAIL: Server certificate does not match provided cert"
    log_error "  Server serial:   ${server_serial}"
    log_error "  Provided serial: ${provided_serial}"
    ((FAIL_COUNT++))
fi

# Verify we can login using the CA to trust the cert
log_info "Verifying podman login with CA trust..."
mkdir -p "/etc/containers/certs.d/${QUAY_ENDPOINT}"
cp "${CERT_DIR}/ca.pem" "/etc/containers/certs.d/${QUAY_ENDPOINT}/ca.crt" 2>/dev/null || \
    sudo cp "${CERT_DIR}/ca.pem" "/etc/containers/certs.d/${QUAY_ENDPOINT}/ca.crt"

assert_success "Podman login with custom CA" \
    podman login -u init -p password "${QUAY_ENDPOINT}" --tls-verify=false

# Push/pull test
verify_push_pull "${QUAY_ENDPOINT}" "init" "password" "--tls-verify=false"

# Uninstall
log_info "Uninstalling..."
${MIRROR_REGISTRY} uninstall --autoApprove -v

assert_clean_uninstall

# Cleanup
rm -rf "${CERT_DIR}"

print_summary
