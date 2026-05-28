#!/usr/bin/env bash
# P1-1: Test installation with custom quayHostname that differs from targetHostname.
#
# Installs with --quayHostname set to a value different from the target host's
# actual hostname. Verifies SERVER_HOSTNAME in config.yaml is correct and
# Quay is accessible at the custom hostname.

set -euo pipefail
source "$(dirname "$0")/lib.sh"

MIRROR_REGISTRY="${1:-.}/mirror-registry"
ACTUAL_HOSTNAME="${QUAY_HOSTNAME:-$(hostname -f)}"
CUSTOM_HOSTNAME="custom-registry.local"
CUSTOM_ENDPOINT="${CUSTOM_HOSTNAME}:8443"

log_info "=== Test: Custom quayHostname Install ==="
log_info "Actual hostname: ${ACTUAL_HOSTNAME}"
log_info "Custom quayHostname: ${CUSTOM_ENDPOINT}"

# Add custom hostname to /etc/hosts pointing to localhost
echo "127.0.0.1 ${CUSTOM_HOSTNAME}" | sudo tee -a /etc/hosts

# Install with custom quayHostname
log_info "Installing with custom quayHostname..."
${MIRROR_REGISTRY} install -v \
    --quayHostname "${CUSTOM_ENDPOINT}" \
    --initPassword password

wait_for_quay "${CUSTOM_ENDPOINT}"
assert_quay_healthy "${CUSTOM_ENDPOINT}"

# Verify config.yaml has the custom hostname
config_hostname=$(grep "SERVER_HOSTNAME" ~/quay-install/quay-config/config.yaml 2>/dev/null | head -1 || echo "")
assert_contains "SERVER_HOSTNAME set to custom hostname" "${config_hostname}" "${CUSTOM_ENDPOINT}"

# Verify Quay is accessible at custom hostname
assert_success "Quay accessible at custom hostname" \
    curl -sk --connect-timeout 10 "https://${CUSTOM_ENDPOINT}/health/instance"

# Verify login works with custom hostname
assert_success "Podman login at custom hostname" \
    podman login -u init -p password "${CUSTOM_ENDPOINT}" --tls-verify=false

# Push/pull with custom hostname
verify_push_pull "${CUSTOM_ENDPOINT}" "init" "password" "--tls-verify=false"

# Uninstall
log_info "Uninstalling..."
${MIRROR_REGISTRY} uninstall --autoApprove -v

assert_clean_uninstall

print_summary
