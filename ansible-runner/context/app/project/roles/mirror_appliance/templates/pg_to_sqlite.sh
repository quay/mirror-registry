#!/bin/bash

# Function that takes postgres data-only dump as "input_file" arg and converts it into a sqlite compatible .sql file
pg_to_sqlite() {
    local input_file="$1"
    local output_file="$2"

    # Read the input file
    sql=$(<"$input_file")

    # Replace PostgreSQL-specific data types with SQLite equivalents
    sql=$(echo "$sql" | sed "s/'true'/1/g")
    sql=$(echo "$sql" | sed "s/'false'/0/g")

    # Remove PostgreSQL-specific commands not supported by SQLite
    sql=$(echo "$sql" | sed -E '/SET\s+\w+\s*=\s*[^;]+;/d')
    sql=$(echo "$sql" | sed -E '/^\s*--.*$/d')
    sql=$(echo "$sql" | sed -E '/ALTER TABLE .*? DISABLE TRIGGER ALL;/d')
    sql=$(echo "$sql" | sed -E '/ALTER TABLE .*? ENABLE TRIGGER ALL;/d')
    sql=$(echo "$sql" | sed -E "s/SELECT pg_catalog\.set_config\('search_path', '', false\);//g")
    sql=$(echo "$sql" | sed -E '/SET SESSION AUTHORIZATION DEFAULT;/d')
    sql=$(echo "$sql" | sed -E '/SET client_encoding = '\''UTF8'\'';/d')

    # Remove lines starting with "pg_dump:"
    sql=$(echo "$sql" | sed -E '/^pg_dump:.*$/d')

    # Remove original PostgreSQL sequence set statements
    sql=$(echo "$sql" | sed -E "s/SELECT pg_catalog\.setval\('public\..*?\', [0-9]+, (true|false)\);//g")

    # Remove the `public.` schema prefix from table names
    sql=$(echo "$sql" | sed "s/INSERT INTO public\./INSERT INTO /g")

    # Remove blank lines
    sql=$(echo "$sql" | sed '/^$/d')

    # Write the output to the specified file
    echo "$sql" > "$output_file"
}

# Check if the script receives two arguments
if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <input_file> <output_file>"
    exit 1
fi

input_file="$1"
output_file="$2"

# Call the function
pg_to_sqlite "$input_file" "$output_file"
