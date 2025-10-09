# Docker Configuration for VolunteerSync

This directory contains Docker configuration files for running the VolunteerSync platform in both development and production environments.

## Files Overview

- **docker-compose.yml** - Development environment configuration
- **docker-compose.prod.yml** - Production environment configuration
- **backend.Dockerfile** - Multi-stage Dockerfile for Go backend API
- **frontend.Dockerfile** - Multi-stage Dockerfile for Next.js frontend

## Development Environment

### Quick Start

From the project root directory:

```bash
# Start all services
docker compose -f docker/docker-compose.yml up

# Start in detached mode
docker compose -f docker/docker-compose.yml up -d

# View logs
docker compose -f docker/docker-compose.yml logs -f

# Stop all services
docker compose -f docker/docker-compose.yml down

# Stop and remove volumes (clean slate)
docker compose -f docker/docker-compose.yml down -v
```

### Services

The development environment includes:

- **PostgreSQL 16** - Database on port 5432
- **Redis 7** - Cache and session store on port 6379
- **Backend API** - Go application on port 8080
- **Frontend** - Next.js application on port 3000

### Default Credentials (Development Only)

**PostgreSQL:**

- Host: localhost
- Port: 5432
- Database: volunteersync
- Username: volunteersync
- Password: volunteersync_dev_password

**Redis:**

- Host: localhost
- Port: 6379
- Password: volunteersync_redis_password

### Accessing Services

- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- Backend Health Check: http://localhost:8080/health
- PostgreSQL: localhost:5432
- Redis: localhost:6379

## Production Environment

### Prerequisites

1. Create a `.env` file in the project root with required environment variables:

```bash
# Database
DB_NAME=volunteersync
DB_USER=volunteersync
DB_PASSWORD=your_secure_password_here

# Redis
REDIS_PASSWORD=your_secure_redis_password_here

# JWT
JWT_SECRET=your_secure_jwt_secret_minimum_32_characters

# CORS
CORS_ALLOWED_ORIGINS=https://yourdomain.com,https://www.yourdomain.com

# API URLs
NEXT_PUBLIC_API_URL=https://api.yourdomain.com/api/v1
NEXT_PUBLIC_APP_URL=https://yourdomain.com
```

### Deployment

```bash
# Start production services
docker compose -f docker/docker-compose.prod.yml up -d

# View logs
docker compose -f docker/docker-compose.prod.yml logs -f

# Stop services
docker compose -f docker/docker-compose.prod.yml down
```

### Production Features

- **Resource Limits**: CPU and memory constraints for each service
- **Health Checks**: All services have health check endpoints
- **Security**: Read-only filesystems, non-root users, no new privileges
- **Logging**: JSON logging with rotation (max 10MB per file)
- **Restart Policy**: Auto-restart on failure
- **SSL/TLS**: Nginx reverse proxy for HTTPS (requires SSL certificates)

### Nginx Reverse Proxy (Optional)

The production compose file includes an optional Nginx service for SSL termination and load balancing. To use it:

1. Create `docker/nginx.conf` with your configuration
2. Place SSL certificates in `docker/ssl/`
3. Uncomment the nginx service in docker-compose.prod.yml

## Building Images

### Backend

```bash
# Build backend image
docker build -f docker/backend.Dockerfile -t volunteersync-backend:latest .

# Build for production
docker build -f docker/backend.Dockerfile --target production -t volunteersync-backend:prod .
```

### Frontend

```bash
# Build frontend for development
docker build -f docker/frontend.Dockerfile --target development -t volunteersync-frontend:dev .

# Build for production
docker build -f docker/frontend.Dockerfile --target production -t volunteersync-frontend:prod .
```

## Volumes

### Development

- `volunteersync-postgres-data` - PostgreSQL data persistence
- `volunteersync-redis-data` - Redis data persistence
- `volunteersync-go-modules` - Go modules cache for faster rebuilds

### Production

- `volunteersync-postgres-data-prod` - PostgreSQL data persistence
- `volunteersync-redis-data-prod` - Redis data persistence
- `volunteersync-nginx-cache-prod` - Nginx cache

## Networking

All services communicate through a dedicated Docker network:

- Development: `volunteersync-network`
- Production: `volunteersync-network-prod`

## Troubleshooting

### Check service health

```bash
# Check all container statuses
docker compose -f docker/docker-compose.yml ps

# Check specific service logs
docker compose -f docker/docker-compose.yml logs backend
docker compose -f docker/docker-compose.yml logs frontend
docker compose -f docker/docker-compose.yml logs postgres
```

### Database connection issues

```bash
# Test PostgreSQL connection
docker compose -f docker/docker-compose.yml exec postgres psql -U volunteersync -d volunteersync

# Check Redis connection
docker compose -f docker/docker-compose.yml exec redis redis-cli -a volunteersync_redis_password ping
```

### Rebuild services

```bash
# Rebuild specific service
docker compose -f docker/docker-compose.yml up --build backend

# Rebuild all services
docker compose -f docker/docker-compose.yml up --build
```

### Clean everything

```bash
# Stop and remove containers, networks, volumes
docker compose -f docker/docker-compose.yml down -v

# Remove all images
docker compose -f docker/docker-compose.yml down --rmi all
```

## Performance Optimization

### Layer Caching

Both Dockerfiles use multi-stage builds and proper layer ordering to maximize cache efficiency:

1. **Backend**: Go modules downloaded before copying source code
2. **Frontend**: Node modules installed before copying source code

### Volume Mounts

Development mode uses volume mounts for hot reloading:

- Backend: Source code mounted, Go modules cached in named volume
- Frontend: Source code mounted, node_modules excluded from mount

## Security Considerations

### Development

- Uses default passwords (NOT for production)
- All ports exposed to host
- Debug logging enabled

### Production

- Requires environment variables for all secrets
- Non-root users in containers
- Read-only filesystems where possible
- Security headers enabled
- TLS encryption recommended (via Nginx)
- No new privileges flag set
- Resource limits enforced

## Next Steps

After Docker setup is complete:

1. Run database migrations: `docker compose exec backend ./migrate.sh up`
2. Verify health checks: `curl http://localhost:8080/health`
3. Access the application: http://localhost:3000
4. Review logs for any errors

## Additional Resources

- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Next.js Docker Deployment](https://nextjs.org/docs/deployment#docker-image)
- [PostgreSQL Docker Official Images](https://hub.docker.com/_/postgres)
- [Redis Docker Official Images](https://hub.docker.com/_/redis)
