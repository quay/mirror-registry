#!/bin/bash
set -euo pipefail

# error tracing
trap 'echo "ERROR at line $LINENO: Command failed - $BASH_COMMAND" >&2; exit 1' ERR

#function that takes postgres data-only dump as "input_file" arg and converts it into a sqlite compatible .sql file
pg_to_sqlite() {
    local input_file="$1"
    local output_file="$2"

    # Validate input file
    [[ ! -f "$input_file" ]] && { echo "Error: Input file '$input_file' not found (line $LINENO)" >&2; exit 1; }
    [[ ! -s "$input_file" ]] && { echo "Error: Input file '$input_file' is empty (line $LINENO)" >&2; exit 1; }

    {
        echo "BEGIN TRANSACTION;"

        # Process conversion
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
        " "$input_file"

        echo "COMMIT;"
        echo "PRAGMA journal_mode=WAL;"
    } > "$output_file"

    # Validate output file
    if [[ ! -f "$output_file" ]]; then
        echo "Error: Output file '$output_file' was not created (line $LINENO)" >&2
        exit 1
    elif [[ ! -s "$output_file" ]]; then
        echo "Error: Output file '$output_file' is empty (line $LINENO)" >&2
        exit 1
    fi
}

pg_to_sqlite "$1" "$2"
echo "Success: Created valid output file '$2'"
