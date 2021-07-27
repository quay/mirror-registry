# Quay Installer

This application will allow user to install Quay and its required components using a simple CLI tool.

## Pre-Requisites

- RHEL 8 or fedora machine with Podman installed
- `sudo` access on desired host (rootless install tbd)
- make (only if compiling using Makefile)

### Compile

To compile the openshift-mirror-registry.tar.gz for distribution, run the following command:

```console
$ git clone https://github.com/quay/openshift-mirror-registry.git
$ cd openshift-mirror-registry
$ make build-online-zip # OR make build-offline-zip
```

This will generate a `openshift-mirror-registry.tar.gz` which contains this README.md, the `openshift-mirror-registry` binary, and the `image-archive.tar` (if using offline installer) which contains the images required to set up Quay.

Once generated, you may untar this file on your desired host machine for installation. You may use the following command:

```console
mkdir openshift-mirror-registry
tar -xzvf openshift-mirror-registry.tar.gz -C openshift-mirror-registry
```

NOTE - With the offline version, this may take some time.

### Installation

Add the following line to host machine `/etc/hosts` file:

```
<target ip>   quay
```

To install Quay on your desired host machine, run the following command:

```console
$ sudo ./openshift-mirror-registry install -v
```

The following flags are also available:

```
--ssh-key   -k  The path of your ssh identity key. This defaults to ~/.ssh/quay_installer
--targetHostname  -H  The hostname of the target you wish to install Quay to. This defaults to localhost.
--targetUsername   -u  The user you wish to ssh into your remote with. This defaults to $USER
```

This command will make the following changes to your machine

- Pulls Quay, Redis, and Postgres containers from quay.io
- Sets up systemd files on host machine to ensure that container runtimes are persistent
- Creates `/etc/quay-install` contains install files, local storage, and config bundle.

### Access Quay

- The Quay console will be accessible at `https://quay:8443`

- Genereate an encrypted password using the password you set up your user with. From the upper right corner, choose `Account Settings` from the dropdown. Then, `Generate Encrypted Password` under `Docker CLI Password`. Then you can download the authorization file or copy the Login Command and paste in your terminal. For example, the following command will login and place authentication in the default location. With `podman` the default location (non-root) is `${XDG_RUNTIME_DIR}/containers/auth.json`. In fedora/RHEL, this is `/run/user/$UID/containers/auth.json` or usually `/run/user/1000/containers/auth.json`:

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
$ sudo ./openshift-mirror-registry uninstall -v
```

This command will delete the `/etc/quay-install` directory and disable all systemd services set up by Quay.
