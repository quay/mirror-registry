import re

# Function that takes postgres data-only dump as "input_file" arg and converts it into a sqlite compatible .sql file
def pg_to_sqlite(input_file, output_file):
    with open(input_file, 'r') as file:
        sql = file.read()

    # Replace PostgreSQL-specific data types with SQLite equivalents
    sql = re.sub(r'\'true\'', '1', sql)
    sql = re.sub(r'\'false\'', '0', sql)

    # Remove PostgreSQL-specific commands not supported by SQLite
    sql = re.sub(r'SET\s+\w+\s*=\s*[^;]+;', '', sql)  # Remove all SET statements
    sql = re.sub(r'--.*', '', sql)  # Remove comments
    sql = re.sub(r'ALTER TABLE .*? DISABLE TRIGGER ALL;', '', sql)  # Remove disable triggers
    sql = re.sub(r'ALTER TABLE .*? ENABLE TRIGGER ALL;', '', sql)  # Remove enable triggers
    sql = re.sub(r"SELECT pg_catalog\.set_config\('search_path', '', false\);", '', sql)  # Remove search path config
    sql = re.sub(r'SET SESSION AUTHORIZATION DEFAULT;', '', sql)  # Remove session authorization
    sql = re.sub(r'SET client_encoding = \'UTF8\';', '', sql)  # Remove client_encoding statement

    # Remove lines starting with "pg_dump:"
    sql = re.sub(r'^pg_dump:.*$', '', sql, flags=re.MULTILINE)

    # Remove original PostgreSQL sequence set statements
    sql = re.sub(r"SELECT pg_catalog\.setval\('public\..*?\', \d+, (true|false)\);", '', sql)

    # Remove the `public.` schema prefix from table names
    sql = re.sub(r'INSERT INTO public\.', 'INSERT INTO ', sql)

    # Remove blank lines
    sql = re.sub(r'\n\s*\n', '\n', sql)
    sql = sql.strip()  # Remove leading and trailing whitespace, including newlines

    with open(output_file, 'w') as file:
        file.write(sql)

if __name__ == "__main__":
    import sys
    if len(sys.argv) != 3:
        print("Usage: python pg_to_sqlite.py <input_file> <output_file>")
        sys.exit(1)

    input_file = sys.argv[1]
    output_file = sys.argv[2]

    pg_to_sqlite(input_file, output_file)

