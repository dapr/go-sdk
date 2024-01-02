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
