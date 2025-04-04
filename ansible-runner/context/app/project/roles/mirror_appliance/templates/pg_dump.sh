#!/bin/bash

CPU_COUNT=$(nproc 2>/dev/null || echo 1)
CPU_COUNT=$(( CPU_COUNT > 0 ? CPU_COUNT : 1))
echo "CPU COUNT: ${CPU_COUNT}"

# clean up postgres tmp/dumpdir if previously created
podman exec quay-postgres rm -rf /tmp/dumpdir

# Take pg_dump of postgres container in parallel and store it in tmp/dumpdir
podman exec -it quay-postgres \
    pg_dump -j "$CPU_COUNT" -Z 0 --no-sync --format=directory \
    --data-only --column-inserts --no-owner --no-privileges \
    --disable-triggers -U user -d quay -f /tmp/dumpdir

# Convert the .dat dump files to a single .sql file using pg_restore
podman exec -i quay-postgres \
    pg_restore -Fd /tmp/dumpdir -f /tmp/pg_data_dump.sql -j "$CPU_COUNT"
