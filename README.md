# Quay Installer

This application will allow user to install Quay and its required components using a simple CLI tool.

## Pre-Requisites

- RHEL 8 machine with Podman installed
- `sudo` access on desired host (rootless install tbd)

## Usage

### Installation

To install Quay on your desired host machine, run the following command:

```
quay-installer install
```

This command will make the following changes to your machine

- Pulls Quay, Redis, and Postgres containers from quay.io
- Sets up systemd files on host machine to ensure that container runtimes are persistent
- Creates `~/quay-install` in `$HOME` which contains install files, local storage, and config bundle.

### Uninstall

To uninstall Quay, run the following command:

```
quay-installer uninstall
```

This command will delete the `~/quay-install` directory and disable all systemd services set up by Quay.
