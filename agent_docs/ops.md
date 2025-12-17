# Operations & Releases

## Build Process

### Prerequisites
- `podman` (or `docker`)
- `make`
- Login to `registry.redhat.io` for image pulls

### Build Commands

```bash
# Online installer (downloads images during install)
make build-online-zip

# Offline installer (includes all images)
make build-offline-zip
```

Both produce `mirror-registry.tar.gz` containing:
- `mirror-registry` binary
- `execution-environment.tar` (Ansible EE)
- `image-archive.tar` (offline only - Quay/Redis images)

### Image Configuration

Image versions are defined in `.env`:
- `QUAY_IMAGE` - Quay container (registry.redhat.io/quay/quay-rhel8)
- `REDIS_IMAGE` - Redis container
- `EE_IMAGE` - Ansible execution environment
- `PAUSE_IMAGE` - Pod pause container
- `SQLITE_IMAGE` - SQLite CLI for database operations

To update images, modify `.env` and rebuild.

## Releases

### Version Format
Tags follow semantic versioning: `v{major}.{minor}.{patch}`

### Creating a Release

1. Create and push a version tag:
   ```bash
   git tag v1.4.0
   git push origin v1.4.0
   ```

2. CI automatically:
   - Builds online and offline installers
   - Creates GitHub release with artifacts
   - Uploads `mirror-registry-online.tar.gz` and `mirror-registry-offline.tar.gz`

### Release Artifacts
- `mirror-registry-online.tar.gz` - Requires internet during install
- `mirror-registry-offline.tar.gz` - Air-gapped installation support
- `README.md`

## Local Development

```bash
# Build Go binary only (for local testing)
make build-golang-executable

# Clean build artifacts
make clean
```

## Container Registry Access

Building requires access to:
- `registry.redhat.io` - Quay, Redis, Ansible EE base images
- `registry.access.redhat.com` - UBI pause image
- `quay.io` - SQLite CLI, mirror-registry-ee

Login before building:
```bash
podman login registry.redhat.io
```
