#!/bin/bash

# Define variables
DB_CONTAINER_NAME="postgres_container"
DB_PORT=5432
LOCAL_PORT=5432
PASSWORD=mysecretpassword

# Run Postgres container
docker run --name $DB_CONTAINER_NAME -e POSTGRES_PASSWORD=$PASSWORD -p $LOCAL_PORT:$DB_PORT -d postgres

# Wait for the container to start
echo "Waiting for Postgres container to start..."
while ! docker exec $DB_CONTAINER_NAME pg_isready -q -h localhost -p $DB_PORT -U postgres; do
    sleep 2
done
echo "Postgres container is now running."

psql "postgresql://postgres:$PASSWORD@localhost:$LOCAL_PORT/postgres"  << EOF
create database hasura with owner postgres;
\c hasura;
create table sagastate ( key text PRIMARY KEY, value text );
GRANT ALL PRIVILEGES ON DATABASE hasura to postgres;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public to postgres;
EOF
echo "psql -h localhost -p $LOCAL_PORT -U postgres -d hasura"
echo "password = $PASSWORD"

# Get the current directory
current_directory=$(pwd)

# Check if the current directory ends with '/test_clients'
if [[ $current_directory == */test_clients ]]; then
    # Remove '/test_clients' from the end and append '/components/secrets.json'
    new_directory="${current_directory%/test_clients}/components/secrets.json"

    # File to be modified
    yaml_file="../components/secrets.yaml"

    # Check if the file exists
    if [[ -f $yaml_file ]]; then
        # Update the specified line in the file
        sed -i '' "s|/Users/stevef/dev/sagaexecutor/components/secrets.json|$new_directory|" "$yaml_file"
        
        echo "Updated $yaml_file with the new directory path $new_directory"
    else
        echo "File $yaml_file does not exist."
    fi
else
    echo "Current directory does not end with '/test_clients'"
fi
