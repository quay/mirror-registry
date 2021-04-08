# Quay Ansible Installer

This application will allow user to install Quay and its required components using a ansible playbooks.

## Pre-Requisites

- RHEL 8 or fedora machine with Podman and Ansible installed
- `sudo` access on desired host (rootless install tbd)

### Install Pre-Requisites

To install the necessary prereqs, run the following commands:

```console
$ sudo pip3 install --upgrade pip
$ pip install ansible
$ sudo yum module install -y container-tools
```

### Installation

Add the following line to host machine `/etc/hosts` file:

```
127.0.0.1   quay
```

Run the Ansible playbook to install Quay:

```
ansible-playbook -i hosts.ini site.yaml
```

This command will make the following changes to your machine

- Pulls Quay, Redis, and Postgres containers from quay.io
- Sets up systemd files on host machine to ensure that container runtimes are persistent
- Creates `/etc/quay-install` which contains install files, local storage, and config bundle.

### Access Quay

- The Quay console will be accessible at `http://quay:8080`

- Create a user with a username and password

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
