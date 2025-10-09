# Frontend Dockerfile for VolunteerSync Next.js 15 Application
# Multi-stage build for optimal image size and performance

# Stage 1: Dependencies installation
FROM node:20-alpine AS deps

# Install dependencies only when needed
RUN apk add --no-cache libc6-compat

WORKDIR /app

# Copy package files
COPY frontend/package.json frontend/bun.lock* frontend/package-lock.json* frontend/yarn.lock* ./

# Install dependencies based on the preferred package manager
RUN \
    if [ -f bun.lock ]; then \
    corepack enable && \
    bun install --frozen-lockfile; \
    elif [ -f yarn.lock ]; then \
    yarn --frozen-lockfile; \
    elif [ -f package-lock.json ]; then \
    npm ci; \
    else \
    echo "No lockfile found. Please use npm, yarn, or bun."; \
    exit 1; \
    fi

# Stage 2: Build stage
FROM node:20-alpine AS builder

WORKDIR /app

# Copy dependencies from deps stage
COPY --from=deps /app/node_modules ./node_modules

# Copy all frontend source code
COPY frontend/ ./

# Set environment variables for build
ENV NEXT_TELEMETRY_DISABLED=1
ENV NODE_ENV=production

# Build the Next.js application
RUN \
    if [ -f bun.lock ]; then \
    corepack enable && \
    bun run build; \
    elif [ -f yarn.lock ]; then \
    yarn build; \
    else \
    npm run build; \
    fi

# Stage 3: Production stage
FROM node:20-alpine AS production

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache wget

# Create non-root user for security
RUN addgroup -g 1001 -S nodejs && \
    adduser -S nextjs -u 1001

# Set environment variables
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

# Copy necessary files from builder
COPY --from=builder /app/public ./public
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static

# Change ownership to non-root user
RUN chown -R nextjs:nodejs /app

# Switch to non-root user
USER nextjs

# Expose port
EXPOSE 3000

# Set hostname
ENV HOSTNAME="0.0.0.0"
ENV PORT=3000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3000 || exit 1

# Run the application
CMD ["node", "server.js"]

# Stage 4: Development stage
FROM node:20-alpine AS development

WORKDIR /app

# Install dependencies
RUN apk add --no-cache libc6-compat

# Copy package files
COPY frontend/package.json frontend/bun.lock* frontend/package-lock.json* frontend/yarn.lock* ./

# Install all dependencies (including dev dependencies)
RUN \
    if [ -f bun.lock ]; then \
    corepack enable && \
    bun install; \
    elif [ -f yarn.lock ]; then \
    yarn; \
    elif [ -f package-lock.json ]; then \
    npm install; \
    else \
    echo "No lockfile found. Please use npm, yarn, or bun."; \
    exit 1; \
    fi

# Copy source code
COPY frontend/ ./

# Set environment variables
ENV NEXT_TELEMETRY_DISABLED=1
ENV NODE_ENV=development

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3000 || exit 1

# Run development server
CMD ["npm", "run", "dev"]
