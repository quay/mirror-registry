# Quay Installer

This application will allow user to install Quay and its required components using a simple CLI tool.

## Pre-Requisites

- RHEL 8 machine with Podman installed
- `sudo` access on desired host (rootless install tbd)
- `go` version 1.16 (only required if compiling from source)

## Usage

### Compile

To compile the installer, run the following commands:

```
$ git clone https://github.com/jonathankingfc/quay-aioi.git
$ cd quay-aioi
$ go build -o quay-installer main.go
```

### Installation

To install Quay on your desired host machine, run the following command:

```
$ sudo ./quay-installer install
```

This command will make the following changes to your machine

- Pulls Quay, Redis, and Postgres containers from quay.io
- Sets up systemd files on host machine to ensure that container runtimes are persistent
- Creates `~/quay-install` in `$HOME` which contains install files, local storage, and config bundle. This will generally be in `/root/quay-install`.

### Uninstall

To uninstall Quay, run the following command:

```
$ sudo ./quay-installer uninstall
```

This command will delete the `~/quay-install` directory and disable all systemd services set up by Quay.

### To Do

- Switch from --net=host to a bridge network (this is safer)
- Figure out SELinux issues (currently not working with SELinux)
- Better config generation with secure passwords
