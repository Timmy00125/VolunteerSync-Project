#!/bin/bash

# VolunteerSync Database Migration Script
# Uses golang-migrate to manage PostgreSQL schema migrations
#
# Usage:
#   ./scripts/migrate.sh up          # Apply all pending migrations
#   ./scripts/migrate.sh down        # Rollback last migration
#   ./scripts/migrate.sh down N      # Rollback N migrations
#   ./scripts/migrate.sh force V     # Force database to version V (use with caution)
#   ./scripts/migrate.sh version     # Show current migration version
#   ./scripts/migrate.sh create NAME # Create new migration files

set -e

# Load environment variables from .env if it exists
if [ -f .env ]; then
  export $(cat .env | grep -v '^#' | xargs)
fi

# Default database connection parameters
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-volunteersync}
DB_PASSWORD=${DB_PASSWORD:-volunteersync_dev}
DB_NAME=${DB_NAME:-volunteersync}
DB_SSLMODE=${DB_SSLMODE:-disable}

# Construct database URL
DATABASE_URL="postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}"

# Migration directory
MIGRATION_DIR="./migrations"

# Check if migrate is installed
if ! command -v migrate &> /dev/null; then
  echo "Error: golang-migrate is not installed"
  echo "Install it using:"
  echo "  macOS: brew install golang-migrate"
  echo "  Linux: See https://github.com/golang-migrate/migrate/tree/master/cmd/migrate#installation"
  echo "  Or use Docker: docker run -v $(pwd)/migrations:/migrations --network host migrate/migrate ..."
  exit 1
fi

# Command handler
COMMAND=${1:-up}

case $COMMAND in
  up)
    echo "Applying all pending migrations..."
    migrate -path $MIGRATION_DIR -database "$DATABASE_URL" up
    echo "✓ Migrations applied successfully"
    ;;
  
  down)
    if [ -n "$2" ]; then
      echo "Rolling back $2 migration(s)..."
      migrate -path $MIGRATION_DIR -database "$DATABASE_URL" down $2
    else
      echo "Rolling back last migration..."
      migrate -path $MIGRATION_DIR -database "$DATABASE_URL" down 1
    fi
    echo "✓ Rollback completed successfully"
    ;;
  
  force)
    if [ -z "$2" ]; then
      echo "Error: Version number required for force command"
      echo "Usage: ./scripts/migrate.sh force VERSION"
      exit 1
    fi
    echo "Forcing database version to $2..."
    migrate -path $MIGRATION_DIR -database "$DATABASE_URL" force $2
    echo "✓ Database version forced to $2"
    ;;
  
  version)
    echo "Current migration version:"
    migrate -path $MIGRATION_DIR -database "$DATABASE_URL" version
    ;;
  
  create)
    if [ -z "$2" ]; then
      echo "Error: Migration name required"
      echo "Usage: ./scripts/migrate.sh create MIGRATION_NAME"
      exit 1
    fi
    echo "Creating new migration: $2"
    migrate create -ext sql -dir $MIGRATION_DIR -seq $2
    echo "✓ Migration files created successfully"
    ;;
  
  *)
    echo "Unknown command: $COMMAND"
    echo ""
    echo "Usage: ./scripts/migrate.sh COMMAND [ARGS]"
    echo ""
    echo "Commands:"
    echo "  up              Apply all pending migrations"
    echo "  down [N]        Rollback last migration (or N migrations)"
    echo "  force VERSION   Force database to specific version (use with caution)"
    echo "  version         Show current migration version"
    echo "  create NAME     Create new migration files"
    echo ""
    echo "Environment Variables:"
    echo "  DB_HOST         Database host (default: localhost)"
    echo "  DB_PORT         Database port (default: 5432)"
    echo "  DB_USER         Database user (default: volunteersync)"
    echo "  DB_PASSWORD     Database password (default: volunteersync_dev)"
    echo "  DB_NAME         Database name (default: volunteersync)"
    echo "  DB_SSLMODE      SSL mode (default: disable)"
    echo ""
    echo "Example:"
    echo "  DB_HOST=db ./scripts/migrate.sh up"
    exit 1
    ;;
esac
