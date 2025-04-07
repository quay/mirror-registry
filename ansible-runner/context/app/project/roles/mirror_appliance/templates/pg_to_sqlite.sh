#!/bin/bash
set -euo pipefail  # Strict error handling

# function that takes postgres data-only dump as "input_file" arg and converts it into a sqlite compatible .sql file
pg_to_sqlite() {
    local input_file="$1"
    local output_file="$2"

    [[ ! -s "$input_file" ]] && { echo "Error: Input file is empty!" >&2; exit 1; }

    # Process with sed
    sed -E "
        s/'true'/1/g;
        s/'false'/0/g;
        /SET\s+\w+\s*=\s*[^;]+;/d;
        /^\s*--.*$/d;
        /ALTER TABLE .*? DISABLE TRIGGER ALL;/d;
        /ALTER TABLE .*? ENABLE TRIGGER ALL;/d;
        s/SELECT pg_catalog\.set_config\('search_path', '', false\);//g;
        /SET SESSION AUTHORIZATION DEFAULT;/d;
        /SET client_encoding = '\''UTF8'\'';/d;
        /^pg_dump:.*$/d;
        s/SELECT pg_catalog\.setval\('public\..*?', [0-9]+, (true|false)\);//g;
        s/INSERT INTO public\./INSERT INTO /g;
        /^\s*$/d;
    " "$input_file" > "$output_file"

    [[ ! -s "$output_file" ]] && { echo "Error: Output file is empty!" >&2; exit 1; }
}

pg_to_sqlite "$1" "$2"
