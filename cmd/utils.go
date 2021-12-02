package cmd

// getImageMetadata provides the metadata needed for a corresponding image
func getImageMetadata(app, imageName, archivePath string) string {
	var statement string

	switch app {
	case "ansible":
		statement = `sudo /usr/bin/podman image import \
					--change 'ENV PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin' \
					--change 'ENV HOME=/home/runner' \
					--change 'ENV container=oci' \
					--change 'ENTRYPOINT=["entrypoint"]' \
					--change 'WORKDIR=/runner' \
					--change 'EXPOSE=6379' \
					--change 'VOLUME=/runner' \
					--change 'CMD ["ansible-runner", "run", "/runner"]' \
					- ` + imageName + ` < ` + archivePath
	case "redis":
		statement = `sudo /usr/bin/podman image import \
					--change 'ENV PATH=/opt/app-root/src/bin:/opt/app-root/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin' \
					--change 'ENV container=oci' \
					--change 'ENV STI_SCRIPTS_URL=image:///usr/libexec/s2i' \
					--change 'ENV STI_SCRIPTS_PATH=/usr/libexec/s2i' \
					--change 'ENV APP_ROOT=/opt/app-root' \
					--change 'ENV HOME=/var/lib/redis' \
					--change 'ENV PLATFORM=el8' \
					--change 'ENV REDIS_VERSION=6' \
					--change 'ENV CONTAINER_SCRIPTS_PATH=/usr/share/container-scripts/redis' \
					--change 'ENV REDIS_PREFIX=/usr' \
					--change 'ENTRYPOINT=["container-entrypoint"]' \
					--change 'USER=1001' \
					--change 'WORKDIR=/opt/app-root/src' \
					--change 'EXPOSE=6379' \
					--change 'VOLUME=/var/lib/redis/data' \
					--change 'CMD ["run-redis"]' \
					- ` + imageName + ` < ` + archivePath
	case "postgres":
		statement = `sudo /usr/bin/podman image import \
					--change 'ENV PATH=/opt/app-root/src/bin:/opt/app-root/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin' \
					--change 'ENV STI_SCRIPTS_URL=image:///usr/libexec/s2i' \
					--change 'ENV STI_SCRIPTS_PATH=/usr/libexec/s2i' \
					--change 'ENV APP_ROOT=/opt/app-root' \
					--change 'ENV APP_DATA=/opt/app-root' \
					--change 'ENV HOME=/var/lib/pgsql' \
					--change 'ENV PLATFORM=el8' \
					--change 'ENV POSTGRESQL_VERSION=10' \
					--change 'ENV POSTGRESQL_PREV_VERSION=9.6' \
					--change 'ENV PGUSER=postgres' \
					--change 'ENV CONTAINER_SCRIPTS_PATH=/usr/share/container-scripts/postgresql' \
					--change 'ENTRYPOINT=["container-entrypoint"]' \
					--change 'WORKDIR=/opt/app-root/src' \
					--change 'EXPOSE=5432' \
					--change 'USER=26' \
					--change 'CMD ["run-postgresql"]' \
					- ` + imageName + ` < ` + archivePath
	case "quay":
		// quay.io
		statement = `sudo /usr/bin/podman image import \
					--change 'ENV PATH=/.local/bin/:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin' \
					--change 'ENV RED_HAT_QUAY=true' \
					--change 'ENV PYTHON_VERSION=3.8' \
					--change 'ENV PYTHON_ROOT=/usr/local/lib/python3.8' \
					--change 'ENV PYTHONUNBUFFERED=1' \
					--change 'ENV PYTHONIOENCODING=UTF-8' \
					--change 'ENV LANG=en_US.utf8' \
					--change 'ENV QUAYDIR=/quay-registry' \
					--change 'ENV QUAYCONF=/quay-registry/conf' \
					--change 'ENV QUAYPATH=.' \
					--change 'ENV container=oci' \
					--change 'ENTRYPOINT=["dumb-init","--","/quay-registry/quay-entrypoint.sh"]' \
					--change 'WORKDIR=/quay-registry' \
					--change 'EXPOSE=7443' \
					--change 'EXPOSE=8080' \
					--change 'EXPOSE=8443' \
					--change 'VOLUME=/conf/stack' \
					--change 'VOLUME=/datastorage' \
					--change 'VOLUME=/tmp' \
					--change 'VOLUME=/var/log' \
					--change 'USER=1001' \
					--change 'CMD ["registry"]' \
					- ` + imageName + ` < ` + archivePath
	}

	return statement
}
