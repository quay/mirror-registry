# Testing & CI/CD

## CI/CD Pipeline

Workflow: `.github/workflows/jobs.yml`

### Triggers
- **Push to version tag** (`v[0-9]+.[0-9]+.[0-9]+`): Full build and release
- **PR with `ok-to-test` label**: Build and run integration tests
- **Manual dispatch**: Release specific version

### Pipeline Stages

1. **Build Installer** - Builds both online and offline installers (matrix: online/offline)
2. **Test Remote Install** - Provisions GCP VM, tests remote SSH-based installation (PRs only)
3. **Test Local Install** - Provisions GCP VM, tests local installation (PRs only)
4. **Publish Release** - Creates GitHub release with artifacts (tags only)

### Concurrency

Pipeline uses `concurrency: group: limit-to-one` to prevent parallel runs.

## Integration Tests

### Requirements

- PRs require the `ok-to-test` label to trigger integration tests
- Tests use Terraform to provision RHEL VMs on Google Cloud

### Test Matrix

Both online and offline installers are tested with:
- **Remote Install**: CLI runs locally, deploys to remote VM via SSH
- **Local Install**: CLI runs on the target VM itself

### Test Scenarios

Each test run verifies:

1. **Fresh Installation**
   - Install Quay with `--initPassword password`
   - Verify Quay is accessible

2. **Image Mirroring**
   - Use oc-mirror to push images to the registry
   - Validates real-world usage

3. **Upgrade**
   - Upgrade the installation
   - Verify services restart correctly

4. **Uninstall**
   - Remove all components with `--autoApprove`

5. **PostgreSQL to SQLite Migration**
   - Install old version (v1.3.10) with PostgreSQL backend
   - Push test image (busybox)
   - Upgrade with new binary (migrates to SQLite)
   - Pull test image to verify data integrity

## Local Testing

### Vagrant Environment

Located in `test/`:
- `Vagrantfile` - Defines test VM
- `Makefile` - Test automation targets

### Running Tests Locally

```bash
cd test

# Run all tests (install, sudo-install, sudo-upgrade)
make

# Individual test targets
make test-install        # Non-root installation test
make test-sudo-install   # Root installation test
make test-sudo-upgrade   # Upgrade from OLD_VERSION (default: 1.2.9)

# Test a specific version
make test-install VERSION=1.3.10

# Keep VM running after test for debugging
make test-install DEBUG=1
```

## Test Infrastructure

### Terraform Contexts

Located in `.github/workflows/`:
- `local-online-installer/` - Local install, online mode
- `local-offline-installer/` - Local install, offline mode
- `remote-online-installer/` - Remote install, online mode
- `remote-offline-installer/` - Remote install, offline mode

### GitHub Actions

Custom actions in `.github/actions/`:
- `setup-terraform/` - Initializes Terraform and provisions VM
- `mirror/` - Runs oc-mirror to test image mirroring
