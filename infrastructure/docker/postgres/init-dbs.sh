#!/bin/bash
set -e

# Initialize multiple databases in PostgreSQL for WAS CRM and Evolution Go
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE DATABASE was_crm;
    CREATE DATABASE evogo_auth;
    CREATE DATABASE evogo_users;
    GRANT ALL PRIVILEGES ON DATABASE was_crm TO $POSTGRES_USER;
    GRANT ALL PRIVILEGES ON DATABASE evogo_auth TO $POSTGRES_USER;
    GRANT ALL PRIVILEGES ON DATABASE evogo_users TO $POSTGRES_USER;
EOSQL
