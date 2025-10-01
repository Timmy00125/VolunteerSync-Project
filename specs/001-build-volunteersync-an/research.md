# Phase 0: Research & Technology Decisions

**Feature**: VolunteerSync Platform  
**Date**: October 1, 2025  
**Status**: Complete

## Executive Summary

This research document consolidates technology decisions, architectural patterns, and best practices for building VolunteerSync. All technical unknowns from the planning phase have been resolved through analysis of modern web development practices, security requirements, and performance constraints.

---

## 1. Frontend Technology Stack

### Decision: Next.js 15 with App Router

**Rationale**:

- **React Server Components (RSC)**: Reduces JavaScript bundle size by rendering non-interactive components on the server, improving performance on slower networks (critical for <2s load time on 3G)
- **App Router**: Provides improved routing, layouts, and data fetching patterns compared to Pages Router
- **Built-in optimization**: Automatic code splitting, image optimization, font optimization
- **TypeScript support**: First-class TypeScript integration for type safety
- **Server Actions**: Simplifies form submissions and mutations without separate API routes
- **SEO-friendly**: Server-side rendering improves SEO for public opportunity pages and organization profiles

**Best Practices**:

- Use Server Components by default for static content (opportunity listings, organization profiles)
- Use Client Components (`'use client'`) only when needed (interactive maps, forms, real-time notifications)
- Implement Progressive Web App (PWA) with next-pwa plugin for offline capabilities and mobile app-like experience
- Use Next.js middleware for authentication and route protection
- Leverage `loading.tsx` and `error.tsx` for consistent UX during data fetching and error handling

**Alternatives Considered**:

- **Vite + React SPA**: Rejected due to lack of built-in SSR, requiring custom server setup
- **Remix**: Strong contender with excellent data loading patterns, but Next.js has larger ecosystem and better documentation
- **SvelteKit**: Excellent performance, but smaller ecosystem and team familiarity with React

---

## 2. UI Component Library & Styling

### Decision: Tailwind CSS + shadcn/ui

**Rationale**:

- **Tailwind CSS**: Utility-first CSS framework enables rapid development and consistent design system
- **shadcn/ui**: Provides accessible, customizable components built on Radix UI primitives
- **Copy-paste approach**: shadcn/ui components are added to codebase (not installed as dependency), allowing full customization
- **Accessibility**: Radix UI primitives provide WCAG 2.1 AA compliance foundation (constitutional requirement)
- **No runtime cost**: Tailwind's purge functionality removes unused CSS, keeping bundles small

**Best Practices**:

- Create custom Tailwind theme extending base configuration (colors, spacing, typography)
- Use shadcn/ui for complex components (dropdowns, dialogs, forms)
- Build custom components for domain-specific UI (opportunity cards, volunteer profiles)
- Implement dark mode support using Tailwind's dark variant
- Use Tailwind's responsive modifiers (sm:, md:, lg:) for mobile-first design

**Alternatives Considered**:

- **Material-UI (MUI)**: Comprehensive but heavyweight; bundle size conflicts with performance goals
- **Chakra UI**: Good accessibility but prescriptive styling harder to customize
- **Plain CSS Modules**: More control but slower development and harder to maintain consistency

---

## 3. Backend Framework & Language

### Decision: Go 1.25 with Gin Web Framework

**Rationale**:

- **Performance**: Go's compiled nature and efficient concurrency model support <500ms API response targets and 10,000 concurrent users
- **Gin framework**: Minimalist, fast HTTP framework with middleware support for auth, logging, rate limiting, CORS
- **Type safety**: Static typing prevents runtime errors and improves maintainability
- **Standard library**: Robust standard library reduces external dependencies
- **Deployment**: Single binary compilation simplifies containerization and deployment
- **Ecosystem**: Mature ecosystem for database access (GORM), JWT handling, testing

**Best Practices**:

- Use structured logging (zerolog or zap) for production-grade observability
- Implement graceful shutdown handling for zero-downtime deployments
- Use context.Context throughout for request-scoped values and cancellation
- Implement middleware chain: logging → recovery → CORS → rate limiting → authentication → authorization
- Use Go modules for dependency management with version pinning
- Follow standard Go project layout (cmd/, internal/, pkg/)

**Alternatives Considered**:

- **Node.js (Express/Fastify)**: JavaScript ecosystem consistency with frontend, but Go's performance and type safety better suit scale requirements
- **Python (FastAPI)**: Excellent developer experience and async support, but Go's performance characteristics better for concurrent load
- **Rust (Actix-web)**: Best performance, but steeper learning curve and longer development time

---

## 4. Architecture Pattern

### Decision: Modular Monolith with Domain-Driven Design

**Rationale**:

- **Module boundaries**: Clear separation by domain (auth, users, organizations, opportunities, etc.) enables parallel development
- **Shared infrastructure**: Modules share database, caching, and authentication layers, avoiding distributed system complexity
- **Future flexibility**: Module boundaries align with potential future microservices extraction if scale demands
- **Development velocity**: Monolith simplifies development, testing, and deployment compared to microservices
- **Constitutional alignment**: Modular architecture supports maintainability principle (Constitution III)

**Module Structure** (each module):

```
module/
├── handlers/      # HTTP request handling (thin layer)
├── services/      # Business logic (core domain logic)
├── repositories/  # Data access (GORM queries)
└── models/        # Domain models (struct definitions)
```

**Best Practices**:

- Each module owns its domain models and business rules
- Modules communicate through well-defined service interfaces (no direct repository access across modules)
- Use dependency injection for testability
- Keep handlers thin: validate input → call service → return response
- Services contain business logic and coordinate repositories
- Repositories handle only data access (CRUD operations, queries)

**Alternatives Considered**:

- **Microservices**: Rejected for V1 due to operational complexity, distributed transaction challenges, and unnecessary overhead for initial scale
- **Layered monolith**: Rejected because organizing by technical layer (controllers, services, repositories) instead of domain creates poor cohesion and coupling across features
- **Event-driven architecture**: Deferred to future phases when async workflows (email campaigns, batch processing) are needed

---

## 5. Database & ORM

### Decision: PostgreSQL 16 with GORM

**Rationale**:

- **PostgreSQL 16**:
  - Mature, reliable ACID-compliant RDBMS
  - Built-in full-text search (avoids external search engine for V1)
  - JSON support for flexible data structures where needed
  - PostGIS extension available if advanced geospatial features needed in future
  - Excellent performance for complex queries with proper indexing
- **GORM**:
  - Idiomatic Go ORM with active ecosystem
  - Prevents SQL injection through parameterized queries (constitutional security requirement)
  - Supports migrations, relationships, eager loading (prevents N+1 queries)
  - Hooks for business logic (BeforeCreate, AfterUpdate, etc.)

**Best Practices**:

- Use migrations for schema management (golang-migrate or GORM AutoMigrate with version control)
- Add database indexes on foreign keys, search columns, and filter fields
- Use GORM's Preload for eager loading to avoid N+1 queries (constitutional performance requirement)
- Implement soft deletes for audit trail (deleted_at timestamp)
- Use database constraints (foreign keys, unique indexes) to enforce data integrity
- Enable PostgreSQL query logging in development to identify slow queries

**Alternatives Considered**:

- **MySQL**: Solid choice, but PostgreSQL's full-text search and JSON support provide more flexibility
- **MongoDB**: NoSQL flexibility not needed; relational model suits volunteer management domain
- **sqlx (raw SQL)**: More control but slower development and more error-prone than ORM

---

## 6. Caching Strategy

### Decision: Redis for Sessions and Application Cache

**Rationale**:

- **Session storage**: JWT refresh tokens, rate limiting counters, session metadata
- **Application cache**: Frequently accessed data (organization profiles, opportunity listings)
- **Performance**: In-memory storage provides <1ms access times
- **Persistence**: Redis RDB snapshots provide durability for session data
- **Expiration**: Built-in TTL support for automatic cache invalidation

**Caching Strategy**:

- **Session tokens**: Store refresh tokens with 7-day TTL
- **Rate limiting**: Store request counters with sliding window algorithm
- **Opportunity listings**: Cache paginated results for 5 minutes
- **Organization profiles**: Cache for 30 minutes, invalidate on update
- **User preferences**: Cache notification settings for 1 hour

**Best Practices**:

- Use cache-aside pattern: check cache → if miss, query DB → store in cache
- Implement cache invalidation on writes (create, update, delete)
- Use Redis keyspaces for organization (sessions:_, cache:opportunities:_, ratelimit:\*)
- Monitor cache hit rate and adjust TTLs based on usage patterns
- Use Redis Sentinel or cluster for high availability in production

**Alternatives Considered**:

- **In-memory map**: Simple but doesn't scale across multiple server instances
- **Memcached**: Simpler than Redis but lacks data structures and persistence

---

## 7. Authentication & Authorization

### Decision: JWT with Refresh Token Rotation + RBAC Middleware

**Rationale**:

- **JWT Access Tokens**: Short-lived (15 minutes), stateless, include user ID and role
- **Refresh Tokens**: Long-lived (7 days), stored in Redis, enable token renewal without re-authentication
- **Token Rotation**: Each refresh generates new access + refresh token pair, invalidating old refresh token (prevents token reuse attacks)
- **RBAC**: Role-based access control with four roles (Super Admin, Org Admin, Coordinator, Volunteer)
- **Argon2**: Memory-hard password hashing algorithm resistant to GPU cracking (stronger than bcrypt)

**Authentication Flow**:

1. User logs in with email/password
2. Server validates credentials, generates access token (15m) + refresh token (7d)
3. Refresh token stored in Redis with user ID
4. Client stores access token in memory, refresh token in httpOnly cookie
5. On access token expiry, client uses refresh token to get new pair
6. Refresh token rotation: old refresh token invalidated, new one issued

**Authorization Middleware**:

```
Request → ExtractJWT → ValidateToken → LoadUserRole → CheckPermission → Handler
```

**Best Practices**:

- Store access tokens in memory (not localStorage due to XSS risk)
- Store refresh tokens in httpOnly, secure, SameSite cookies
- Implement token blacklist in Redis for immediate logout/revocation
- Rate limit login attempts (5 attempts per 15 minutes per IP)
- Log all authentication events (login, failed attempts, token refresh)
- Use HTTPS in production (TLS 1.3+) to protect tokens in transit

**Alternatives Considered**:

- **Session-based auth**: Requires server-side session storage for every request, doesn't scale horizontally as easily
- **OAuth 2.0 providers**: Deferred to V2 (spec excludes social login for V1)
- **Paseto tokens**: More secure than JWT but less ecosystem support

---

## 8. State Management (Frontend)

### Decision: TanStack Query for Server State + Zustand for Client State

**Rationale**:

- **TanStack Query (React Query)**:
  - Handles server state (API data fetching, caching, synchronization)
  - Automatic background refetching, optimistic updates, cache invalidation
  - Reduces boilerplate compared to manual fetch + useState
  - Built-in loading/error states
- **Zustand**:
  - Minimal client state management (UI state, user preferences, temporary form data)
  - Simpler than Redux, no boilerplate
  - Works seamlessly with React hooks

**State Boundaries**:

- **Server state** (TanStack Query): Opportunities, organizations, user profiles, volunteer hours, notifications
- **Client state** (Zustand): UI state (modal open/closed, selected filters, draft form data)
- **URL state** (Next.js router): Search filters, pagination, active tab (enables shareable URLs)

**Best Practices**:

- Use TanStack Query for all API data fetching
- Configure appropriate staleTime and cacheTime per query
- Implement optimistic updates for mutations (hours logging, registrations)
- Use Zustand for simple UI state that doesn't need persistence
- Store user preferences in client state with localStorage persistence

**Alternatives Considered**:

- **Redux Toolkit**: More powerful but overkill for this application; TanStack Query handles most state needs
- **Recoil**: Interesting atom-based state, but Zustand + TanStack Query is simpler
- **React Context**: Suitable for simple apps but doesn't provide caching, refetching, or optimistic updates

---

## 9. Form Handling & Validation

### Decision: React Hook Form + Zod

**Rationale**:

- **React Hook Form**: Performant form library using uncontrolled components, reduces re-renders
- **Zod**: TypeScript-first schema validation library
- **Integration**: zod schemas provide both runtime validation and TypeScript types
- **Bundle size**: React Hook Form is lightweight (~9KB), Zod is ~14KB

**Validation Strategy**:

- Define Zod schemas for all forms (registration, opportunity creation, profile updates)
- Use zodResolver to integrate Zod with React Hook Form
- Client-side validation provides immediate feedback
- Server-side validation enforces security (never trust client)
- Share validation schemas between frontend and backend where possible (using TypeScript codegen or manual sync)

**Best Practices**:

- Create reusable Zod schemas in `lib/validations/`
- Use React Hook Form's Controller for complex inputs (date pickers, rich text editors)
- Implement field-level validation for real-time feedback
- Display validation errors clearly with accessible error messages
- Use server response errors to update form state (email already exists, etc.)

**Alternatives Considered**:

- **Formik**: Popular but heavier and slower than React Hook Form
- **Plain React state**: Too much boilerplate for complex forms with validation

---

## 10. Testing Strategy

### Decision: Multi-Layer Testing with TDD Approach

**Frontend Testing**:

- **Unit Tests (Jest)**: Pure functions, utilities, validation logic
- **Component Tests (React Testing Library)**: UI component behavior, user interactions
- **E2E Tests (Playwright)**: Critical user flows (registration, opportunity search, event signup)

**Backend Testing**:

- **Unit Tests (Go testing + Testify)**: Services, utilities, business logic
- **HTTP Tests (httptest)**: Handler tests with mocked services
- **Integration Tests (testcontainers)**: Full stack tests with real PostgreSQL instance

**TDD Workflow** (Constitutional requirement):

1. Write failing test for new feature/requirement
2. Implement minimum code to pass test
3. Refactor code while keeping tests green
4. Repeat

**Coverage Targets**:

- Minimum 80% code coverage (constitutional requirement)
- Critical paths (authentication, registration, hours logging) should have 95%+ coverage
- Use coverage reports to identify untested code paths

**Best Practices**:

- Write contract tests first (OpenAPI spec validation)
- Test user behaviors, not implementation details
- Use test fixtures for consistent test data
- Mock external dependencies (Redis, external APIs)
- Run tests in CI pipeline before merge

**Alternatives Considered**:

- **Cypress**: Good for E2E but Playwright has better performance and multi-browser support
- **Manual testing only**: Rejected due to constitutional requirement for automated testing and 80% coverage

---

## 11. Maps & Geospatial Features

### Decision: Leaflet for Interactive Maps

**Rationale**:

- **Open-source**: No licensing costs, extensive plugin ecosystem
- **Performance**: Lightweight (~140KB), renders efficiently
- **Customization**: Full control over map styling and markers
- **Mobile-friendly**: Touch interactions for mobile devices
- **Integration**: Works seamlessly with React (react-leaflet wrapper)

**Features**:

- Display opportunities on interactive map with clustered markers
- Filter opportunities by distance radius (5, 10, 25, 50 miles)
- Show organization locations
- Geocoding for address input (use OpenStreetMap Nominatim or Mapbox Geocoding API)

**Best Practices**:

- Lazy load map component (Next.js dynamic import) to reduce initial bundle size
- Use marker clustering for dense areas (Leaflet.markercluster plugin)
- Implement geospatial queries in PostgreSQL (PostGIS extension if needed, or simple lat/lon distance calculation)
- Cache geocoded addresses in database to avoid repeated API calls

**Alternatives Considered**:

- **Google Maps**: Excellent but requires API key and has usage costs
- **Mapbox**: Beautiful maps but also has usage costs
- **OpenStreetMap only**: Free but Leaflet provides better interactivity

---

## 12. Analytics & Charting

### Decision: Recharts for Data Visualization

**Rationale**:

- **React-based**: Native React components for easy integration
- **Declarative**: Simple API for common chart types (line, bar, pie, area)
- **Responsive**: Automatically adjusts to container size
- **Bundle size**: ~100KB, acceptable for dashboard features

**Charts Needed**:

- **Organization Dashboard**: Volunteer hours over time (line chart), event fill rates (bar chart), volunteer retention (pie chart)
- **Volunteer Dashboard**: Personal hours over time (area chart), organizations supported (pie chart)
- **Admin Dashboard**: Platform growth (line chart), engagement metrics (bar chart)

**Best Practices**:

- Lazy load chart components (only on dashboard pages)
- Use consistent color palette from Tailwind theme
- Implement loading skeletons while data fetches
- Make charts accessible (provide data tables as alternative view)

**Alternatives Considered**:

- **Chart.js**: More features but not React-native (requires react-chartjs-2 wrapper)
- **D3.js**: Most powerful but overkill and steep learning curve for basic charts
- **Victory**: Good React integration but larger bundle size

---

## 13. Real-Time Notifications

### Decision: In-Platform Notifications with Optional Browser Push

**Rationale**:

- **In-platform notifications**: Primary notification system, displayed in notification center
- **Browser push notifications**: Optional, requires user permission
- **Polling approach**: Frontend polls `/api/notifications/unread` every 30 seconds when user is active
- **WebSocket deferred**: WebSocket for real-time push deferred to V2 to reduce complexity

**Notification Flow**:

1. Backend event triggers notification creation (registration, hours logged, message received)
2. Notification stored in PostgreSQL
3. Frontend polls for unread notifications
4. User sees unread count in header
5. Clicking notification marks as read

**Best Practices**:

- Use TanStack Query's refetchInterval for polling
- Implement exponential backoff if API errors occur
- Pause polling when tab is inactive (Page Visibility API)
- Batch notifications to avoid spam (group by type and time)
- Provide notification preferences page for users to control frequency

**Alternatives Considered**:

- **WebSocket**: Real-time but adds complexity for server scaling (need sticky sessions or Redis pub/sub)
- **Server-Sent Events (SSE)**: Simpler than WebSocket but limited browser support
- **Email/SMS**: Excluded from V1 per feature spec clarifications

---

## 14. Development Environment & Tooling

### Decision: Docker Compose for Local Development

**Rationale**:

- **Consistency**: Identical environment for all developers
- **Services orchestration**: PostgreSQL, Redis, backend, frontend run together
- **Hot reload**:
  - Backend: Use Air for Go hot reload
  - Frontend: Next.js dev server built-in hot reload
- **Easy onboarding**: Single `docker-compose up` command starts entire stack

**docker-compose.yml Services**:

- `postgres`: PostgreSQL 16 with persistent volume
- `redis`: Redis 7 for caching and sessions
- `backend`: Go API with Air hot reload, mounted source code
- `frontend`: Next.js dev server with mounted source code
- `nginx`: Reverse proxy (optional for production-like setup)

**Best Practices**:

- Use environment variables for configuration (`.env` file)
- Mount source code volumes for hot reload
- Use named volumes for persistent data (PostgreSQL, Redis)
- Expose ports for direct service access during debugging
- Provide seed data scripts for local development

**Alternatives Considered**:

- **Local installation**: Harder to maintain consistency across team members
- **Kubernetes for local dev**: Overkill, too complex for development environment

---

## 15. CI/CD Pipeline

### Decision: GitHub Actions for CI/CD

**Rationale**:

- **Native GitHub integration**: Seamless with repository
- **Free for public repos**: Cost-effective for open-source
- **Matrix builds**: Test multiple versions (Go 1.25, Node 20)
- **Caching**: Cache dependencies to speed up builds

**Pipeline Stages**:

**CI Pipeline (on push, PR)**:

1. Lint: Run golangci-lint (backend), ESLint + Prettier (frontend)
2. Test: Run unit tests (Go, Jest), integration tests, E2E tests (Playwright)
3. Coverage: Upload coverage reports, enforce 80% minimum
4. Security: Run Snyk or Dependabot for vulnerability scanning
5. Build: Compile Go binary, build Next.js production bundle

**CD Pipeline (on merge to main)**:

1. Build Docker images (backend, frontend)
2. Push images to container registry
3. Deploy to staging environment
4. Run smoke tests
5. Deploy to production (manual approval gate)

**Best Practices**:

- Run tests in parallel where possible
- Use GitHub Actions cache for dependencies
- Fail fast on linting errors
- Require passing CI before merge
- Use secrets for API keys and credentials

**Alternatives Considered**:

- **GitLab CI**: Good but requires GitLab
- **CircleCI**: Excellent but costs more than GitHub Actions
- **Jenkins**: Too complex for this project's needs

---

## 16. Security Best Practices

### Password Security

- **Argon2id** for password hashing (winner of Password Hashing Competition)
- Salt rounds: 3 iterations, 64MB memory, 4 parallelism
- Never log passwords or include in error messages

### Input Validation

- Validate all inputs on both client and server
- Sanitize user-generated content before rendering (prevent XSS)
- Use parameterized queries (GORM) to prevent SQL injection
- Implement rate limiting on all public endpoints

### Headers & Middleware

- CORS configuration in Gin (restrict allowed origins)
- Security headers: X-Frame-Options, X-Content-Type-Options, CSP
- HSTS header for HTTPS enforcement

### Secrets Management

- Use environment variables for secrets (never commit to git)
- Rotate JWT signing keys periodically
- Use separate secrets for dev/staging/production

### Monitoring

- Log all authentication events (login, logout, failed attempts)
- Monitor for suspicious patterns (rapid requests, failed logins)
- Set up alerts for security anomalies

---

## 17. Deployment Strategy

### Decision: Docker Containers with Docker Compose (Production)

**Rationale**:

- **Containerization**: Consistent runtime environment across dev/staging/production
- **Orchestration**: Docker Compose for simple multi-container deployment
- **Scalability**: Can migrate to Kubernetes if horizontal scaling needed in future

**Production Stack**:

- **Backend**: Go API container (multi-stage build for small image size)
- **Frontend**: Next.js container (standalone output mode)
- **Database**: Managed PostgreSQL (AWS RDS, DigitalOcean Managed Database, or self-hosted with backups)
- **Cache**: Managed Redis or self-hosted Redis container
- **Reverse Proxy**: Nginx for SSL termination, static file serving, load balancing

**Best Practices**:

- Use multi-stage Docker builds to minimize image size
- Run containers as non-root user
- Implement health check endpoints (`/health`, `/ready`)
- Use Docker secrets for sensitive data
- Set resource limits (memory, CPU) for containers

**Deployment Options** (in order of complexity):

1. **Single VPS**: DigitalOcean droplet or AWS EC2 with Docker Compose (suitable for V1)
2. **Platform-as-a-Service**: Railway, Render, or Fly.io (easiest but less control)
3. **Kubernetes**: GKE, EKS, or self-managed (overkill for V1, consider for scale)

---

## 18. Internationalization (Future Consideration)

**V1 Status**: English only per feature spec

**V2 Preparation**:

- Use Next.js i18n features for future multi-language support
- Structure content to be externalizable (no hardcoded strings in components)
- Use `next-intl` or `react-i18next` libraries
- Database schema supports future language fields (opportunity title/description in multiple languages)

---

## 19. Accessibility Compliance (WCAG 2.1 Level AA)

**Requirements**:

- Semantic HTML elements (nav, main, article, aside)
- ARIA labels for interactive elements
- Keyboard navigation support (tab order, focus indicators)
- Color contrast ratios (4.5:1 for normal text, 3:1 for large text)
- Alt text for all images
- Form labels and error associations
- Skip links for navigation

**Tools**:

- axe DevTools for automated testing
- Lighthouse accessibility audits
- Manual keyboard navigation testing
- Screen reader testing (NVDA, JAWS, VoiceOver)

**Best Practices**:

- Use shadcn/ui components (built on Radix UI with accessibility)
- Test with keyboard-only navigation
- Include accessibility in code reviews
- Run automated accessibility tests in CI

---

## 20. Performance Optimization

### Frontend Optimizations

- **Code splitting**: Automatic via Next.js App Router
- **Image optimization**: Next.js Image component with WebP format
- **Font optimization**: next/font for optimized font loading
- **Lazy loading**: Dynamic imports for large components (maps, charts)
- **Bundle analysis**: Use @next/bundle-analyzer to identify bloat

### Backend Optimizations

- **Database indexing**: Add indexes on foreign keys and frequently queried columns
- **Connection pooling**: GORM connection pool (max 10 connections)
- **Query optimization**: Use GORM Preload to avoid N+1 queries
- **Response compression**: Gzip middleware for API responses
- **Pagination**: Limit result sets to 20-50 items per page

### Monitoring

- Use Lighthouse for frontend performance metrics
- Use Go pprof for backend profiling
- Monitor API response times in production
- Set up alerts for slow queries

---

## Research Status: COMPLETE ✅

All technical unknowns have been resolved. Technology stack, architectural patterns, best practices, and implementation strategies are documented. Ready to proceed to Phase 1: Design & Contracts.

**Key Deliverables**:

- Frontend: Next.js 15, TypeScript, Tailwind CSS, shadcn/ui
- Backend: Go 1.25, Gin, modular monolith with DDD
- Database: PostgreSQL 16 with GORM
- Caching: Redis
- Testing: Jest, React Testing Library, Playwright, Go testing, 80% coverage
- Infrastructure: Docker, Docker Compose, GitHub Actions
- Security: JWT with refresh tokens, Argon2, RBAC, rate limiting

**Next Phase**: Generate data models, API contracts, and integration test scenarios.
