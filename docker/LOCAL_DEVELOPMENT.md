# Local Development Configuration

## Port Mappings

The Docker Compose development environment uses **non-standard ports** to avoid conflicts with locally installed PostgreSQL and Redis/Valkey services.

### Host Port Mappings (connecting from your machine)

When connecting to Docker services from your **host machine** (e.g., using a database GUI, Redis CLI, or running backend locally):

| Service     | Host Port | Container Port | Connection String Example                                                            |
| ----------- | --------- | -------------- | ------------------------------------------------------------------------------------ |
| PostgreSQL  | **5433**  | 5432           | `postgresql://volunteersync:volunteersync_dev_password@localhost:5433/volunteersync` |
| Redis       | **6380**  | 6379           | `redis://localhost:6380` (password: `volunteersync_redis_password`)                  |
| Backend API | 8080      | 8080           | `http://localhost:8080`                                                              |
| Frontend    | 3000      | 3000           | `http://localhost:3000`                                                              |

### Container-to-Container Communication

When services communicate **within the Docker network** (e.g., backend connecting to postgres/redis):

| Service    | Hostname   | Port | Used By           |
| ---------- | ---------- | ---- | ----------------- |
| PostgreSQL | `postgres` | 5432 | Backend container |
| Redis      | `redis`    | 6379 | Backend container |

## Why Different Ports?

The host ports (5433, 6380) are mapped to avoid conflicts with:

- Local PostgreSQL typically running on port 5432
- Local Redis/Valkey typically running on port 6379

This allows you to run both local services and Docker services simultaneously.

## Connecting to Services

### From Your Host Machine

```bash
# PostgreSQL
psql -h localhost -p 5433 -U volunteersync -d volunteersync
# Password: volunteersync_dev_password

# Redis
redis-cli -h localhost -p 6380 -a volunteersync_redis_password

# Backend API Health Check
curl http://localhost:8080/health

# Frontend
open http://localhost:3000
```

### Running Backend Locally (outside Docker)

If you want to run the backend locally (not in Docker) but connect to Dockerized PostgreSQL and Redis:

```bash
cd backend

# Set environment variables
export DB_HOST=localhost
export DB_PORT=5433  # ← Note: 5433, not 5432
export DB_USER=volunteersync
export DB_PASSWORD=volunteersync_dev_password
export DB_NAME=volunteersync
export DB_SSLMODE=disable

export REDIS_HOST=localhost
export REDIS_PORT=6380  # ← Note: 6380, not 6379
export REDIS_PASSWORD=volunteersync_redis_password
export REDIS_DB=0

# Run migrations
./scripts/migrate.sh up

# Run backend
go run cmd/api/main.go
```

### Running Frontend Locally (outside Docker)

```bash
cd frontend

# Set environment variables
export NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
export NEXT_PUBLIC_APP_URL=http://localhost:3000

# Install dependencies
npm install

# Run development server
npm run dev
```

## Common Issues

### "Address already in use" errors

If you see port binding errors:

1. **Check what's using the ports:**

   ```bash
   sudo netstat -tulpn | grep -E ':(5432|5433|6379|6380)'
   # or
   ss -tulpn | grep -E ':(5432|5433|6379|6380)'
   ```

2. **Stop local services if needed:**

   ```bash
   sudo systemctl stop postgresql
   sudo systemctl stop redis-server
   sudo systemctl stop valkey-server
   ```

3. **Remove old Docker containers:**
   ```bash
   docker compose -f docker/docker-compose.yml down
   docker container prune
   ```

### Database Connection Issues

- **Inside Docker containers**: Use `postgres:5432` and `redis:6379`
- **From host machine**: Use `localhost:5433` and `localhost:6380`
- **Check container health**: `docker ps` (look for "healthy" status)

### Migration Issues

Make sure you're using the correct port:

```bash
# If running migrations from host against Docker database:
export DB_PORT=5433
./scripts/migrate.sh up

# If running migrations inside Docker network:
docker exec -it volunteersync-backend ./scripts/migrate.sh up
```

## Quick Start

```bash
# Start all services
docker compose -f docker/docker-compose.yml up -d

# Check status
docker compose -f docker/docker-compose.yml ps

# View logs
docker compose -f docker/docker-compose.yml logs -f backend

# Stop all services
docker compose -f docker/docker-compose.yml down

# Clean slate (remove volumes)
docker compose -f docker/docker-compose.yml down -v
```
