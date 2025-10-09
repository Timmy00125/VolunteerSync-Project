# Database Migrations

This directory contains SQL migration files for the VolunteerSync database schema.

## Migration Tool

We use [golang-migrate](https://github.com/golang-migrate/migrate) for database migrations.

### Installation

**macOS:**

```bash
brew install golang-migrate
```

**Linux:**
See [official installation guide](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate#installation)

**Docker (alternative):**

```bash
# For local development (host port 5433)
docker run -v $(pwd)/migrations:/migrations --network host migrate/migrate \
  -path /migrations -database "postgresql://user:pass@localhost:5433/dbname?sslmode=disable" up

# For Docker network (container port 5432)
docker run -v $(pwd)/migrations:/migrations --network volunteersync-network migrate/migrate \
  -path /migrations -database "postgresql://user:pass@postgres:5432/dbname?sslmode=disable" up
```

## Using the Migration Script

A convenience script is provided at `scripts/migrate.sh` to simplify running migrations.

### Basic Commands

```bash
# Apply all pending migrations
./scripts/migrate.sh up

# Rollback last migration
./scripts/migrate.sh down

# Rollback N migrations
./scripts/migrate.sh down 3

# Show current migration version
./scripts/migrate.sh version

# Create a new migration
./scripts/migrate.sh create add_user_preferences

# Force database to specific version (use with caution!)
./scripts/migrate.sh force 1
```

### Environment Configuration

Set these environment variables to configure database connection:

- `DB_HOST` - Database host (default: localhost)
- `DB_PORT` - Database port (default: 5432)
- `DB_USER` - Database user (default: volunteersync)
- `DB_PASSWORD` - Database password (default: volunteersync_dev)
- `DB_NAME` - Database name (default: volunteersync)
- `DB_SSLMODE` - SSL mode (default: disable)

### Example with Docker Compose

When using the Docker Compose setup, the database runs on the `db` service:

```bash
DB_HOST=db ./scripts/migrate.sh up
```

Or create a `.env` file in the `backend/` directory:

```env
DB_HOST=db
DB_PORT=5432
DB_USER=volunteersync
DB_PASSWORD=volunteersync_dev
DB_NAME=volunteersync
DB_SSLMODE=disable
```

## Migration File Naming Convention

Migration files follow this pattern:

```
{version}_{description}.up.sql
{version}_{description}.down.sql
```

Example:

- `000001_initial_schema.up.sql`
- `000001_initial_schema.down.sql`

## Current Migrations

### 000001_initial_schema

**Status**: Placeholder (will be populated in Phase 3.3, Task T046-T047)

**Purpose**: Creates the initial database schema including:

- Users table with security questions
- Organizations table with geocoding
- Volunteer profiles with availability
- Opportunities with recurrence support
- Registrations with hours tracking
- Hours logs audit table
- Messages and notifications
- Lookup tables (skills, causes, achievements)
- Junction tables for N:M relationships
- Indexes on foreign keys and frequently queried fields

## Best Practices

1. **Always test migrations locally first** before applying to production
2. **Write reversible migrations** - ensure `down` migrations correctly undo `up` migrations
3. **One logical change per migration** - keep migrations focused and atomic
4. **Use transactions** - wrap DDL statements in transactions when possible
5. **Backup before migrating** - always backup production databases before running migrations
6. **Version control** - commit migration files to git
7. **Never modify existing migrations** - create new migrations to fix issues

## Troubleshooting

### Migration fails with "dirty database"

If a migration fails partially, the database may be in a "dirty" state. Check the current version:

```bash
./scripts/migrate.sh version
```

If it shows "dirty", you need to manually fix the database and then force the version:

```bash
# After manually fixing the database
./scripts/migrate.sh force VERSION_NUMBER
```

### Connection refused

Ensure the database is running and environment variables are set correctly. For Docker Compose:

```bash
docker-compose ps  # Check if db service is running
docker-compose logs db  # Check database logs
```

## References

- [golang-migrate documentation](https://github.com/golang-migrate/migrate)
- [PostgreSQL migration best practices](https://www.postgresql.org/docs/current/ddl.html)
