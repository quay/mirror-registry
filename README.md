# Quay Installer

This application will allow user to install Quay and its required components using a simple CLI tool.

## Pre-Requisites

- RHEL 8 or fedora machine with Podman installed
- `sudo` access on desired host (rootless install tbd)
- make (only if compiling using Makefile)

### Compile

To compile the quay-installer.tar.gz for distribution, run the following command:

```console
$ git clone https://github.com/jonathankingfc/quay-aioi.git
$ cd quay-aioi
$ make build-online-zip # OR make build-offline-zip
```

This will generate a `quay-installer.tar.gz` which contains this README.md, the `quay-installer` binary, and the `image-archive.tar` (if using offline installer) which contains the images required to set up Quay.

Once generated, you may untar this file on your desired host machine for installation. You may use the following command:

```console
tar -xzvf quay-installer.tar.gz
```

NOTE - With the offline version, this may take some time.

### Installation

Add the following line to host machine `/etc/hosts` file:
```
127.0.0.1   quay
```

To install Quay on your desired host machine, run the following command:

```console
$ sudo ./quay-installer install -v
```

This command will make the following changes to your machine

- Pulls Quay, Redis, and Postgres containers from quay.io
- Sets up systemd files on host machine to ensure that container runtimes are persistent
- Creates `~/quay-install` in `$HOME` which contains install files, local storage, and config bundle. This will generally be in `/root/quay-install`.

### Access Quay

*  The Quay console will be accessible at `http://quay:8080`

*  Create a user with a username and password

*  Genereate an encrypted password using the password you set up your user with. From the upper right corner, choose `Account Settings` from the dropdown. Then, `Generate Encrypted Password` under `Docker CLI Password`. Then you can download the authorization file or copy the Login Command and paste in your terminal.  For example, the following command will login and place authentication in the default location. With `podman` the default location (non-root) is `${XDG_RUNTIME_DIR}/containers/auth.json`. In fedora/RHEL, this is `/run/user/$UID/containers/auth.json` or usually `/run/user/1000/containers/auth.json`:
```console
$ podman login -u <username> -p <the encrypted password from quay console> --tls-verify=false quay:8080
```
or, if you download the authfile to a non-default location (--authfile can be added to any podman command):
```console
$ podman login -u <username> -p <the encrypted password from quay console> --tls-verify=false --authfile=~/path/to/authfile quay:8080
```

After logging in, you can run commands such as:
```console
$ podman tag docker.io/library/busybox:latest quay:8080/<username>/busybox:latest
$ podman push quay:8080/<username>/busybox:latest --tls-verify=false
$ podman pull quay:8080/<username>/busybox:latest --tls-verify=false
```

### Uninstall

To uninstall Quay, run the following command:

```console
$ sudo ./quay-installer uninstall -v
```

This command will delete the `~/quay-install` directory and disable all systemd services set up by Quay.

### To Do

- Switch from --net=host to a bridge network (this is safer)
- Better config generation with secure passwords
