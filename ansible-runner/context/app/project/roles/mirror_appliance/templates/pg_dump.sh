#!/bin/bash

CPU_COUNT=$(nproc 2>/dev/null || echo 1)
CPU_COUNT=$(( CPU_COUNT > 0 ? CPU_COUNT : 1))
echo "CPU COUNT: ${CPU_COUNT}"

# clean up postgres tmp/pgdump_dir if previously created
podman exec quay-postgres rm -rf /tmp/pgdump_dir

# Take pg_dump of postgres container in parallel and store it in tmp/pgdump_dir
podman exec -it quay-postgres \
    pg_dump -j "$CPU_COUNT" -Z 0 --no-sync --format=directory \
    --data-only --column-inserts --no-owner --no-privileges \
    --disable-triggers -U user -d quay -f /tmp/pgdump_dir

# Convert the .dat dump files to a single .sql file using pg_restore
podman exec -i quay-postgres \
    pg_restore -Fd /tmp/pgdump_dir -f /tmp/pg_data_dump.sql -j "$CPU_COUNT"
