#!/usr/bin/env bash
# P2-5: Test that install generates and displays a random password
# when --initPassword is not provided.

set -euo pipefail
source "$(dirname "$0")/lib.sh"

MIRROR_REGISTRY="${1:-.}/mirror-registry"
HOSTNAME="${QUAY_HOSTNAME:-$(hostname -f)}"
QUAY_ENDPOINT="${HOSTNAME}:8443"

log_info "=== Test: Auto-Generated Password ==="

# Install without --initPassword
log_info "Installing without --initPassword..."
install_output=$( ${MIRROR_REGISTRY} install -v \
    --quayHostname "${QUAY_ENDPOINT}" 2>&1 )

wait_for_quay "${QUAY_ENDPOINT}"

# The install output should contain the generated password
# Format: "Quay is available at https://host:8443 with credentials (init, <password>)"
if echo "${install_output}" | grep -q "credentials (init,"; then
    log_info "PASS: Install output contains generated credentials"
    ((PASS_COUNT++))

    # Extract the password from output
    generated_password=$(echo "${install_output}" | grep "credentials (init," | sed 's/.*credentials (init, \(.*\))/\1/' | tr -d ')')
    log_info "Generated password length: ${#generated_password}"

    if [[ ${#generated_password} -gt 8 ]]; then
        log_info "PASS: Generated password has sufficient length"
        ((PASS_COUNT++))
    else
        log_error "FAIL: Generated password too short: ${#generated_password} chars"
        ((FAIL_COUNT++))
    fi

    # Verify we can login with the generated password
    if podman login -u init -p "${generated_password}" "${QUAY_ENDPOINT}" --tls-verify=false 2>/dev/null; then
        log_info "PASS: Login succeeded with generated password"
        ((PASS_COUNT++))
    else
        log_error "FAIL: Login failed with generated password"
        ((FAIL_COUNT++))
    fi
else
    log_error "FAIL: Install output does not contain generated credentials"
    log_error "  Output: ${install_output}"
    ((FAIL_COUNT++))
fi

# Cleanup
log_info "Uninstalling..."
${MIRROR_REGISTRY} uninstall --autoApprove -v

print_summary
