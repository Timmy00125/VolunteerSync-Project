# GitHub Actions CI/CD Workflows

This directory contains the automated CI/CD pipelines for the VolunteerSync platform.

## Workflows

### 1. CI Pipeline (`ci.yml`)

**Trigger**: Push to main, feature/_, bugfix/_, and pull requests to main

**Purpose**: Continuous Integration - Ensures code quality and tests pass before merge

**Jobs**:

- **Backend Linting**: Runs `golangci-lint` on Go code
- **Frontend Linting**: Runs ESLint and Prettier on TypeScript/React code
- **Backend Tests**: Unit and integration tests with PostgreSQL and Redis services
  - Enforces 80% minimum test coverage
  - Uploads coverage to Codecov
- **Frontend Tests**: Jest unit tests
  - Enforces 80% minimum test coverage
  - Uploads coverage to Codecov
- **E2E Tests**: Playwright end-to-end tests with full Docker Compose stack
- **Backend Build**: Compiles Go binary for Linux
- **Frontend Build**: Builds Next.js production bundle

**Artifacts**:

- Backend binary (7-day retention)
- Frontend build (.next folder, 7-day retention)
- Playwright test reports (30-day retention)
- Code coverage reports (uploaded to Codecov)

**Caching Strategy**:

- Go modules and build cache
- npm packages
- Docker layer caching for E2E tests

---

### 2. CD Pipeline (`cd.yml`)

**Trigger**: Push to main, manual workflow dispatch

**Purpose**: Continuous Deployment - Builds Docker images and deploys to staging/production

**Jobs**:

- **Build Backend**: Builds and pushes backend Docker image to GitHub Container Registry (GHCR)
- **Build Frontend**: Builds and pushes frontend Docker image to GHCR
- **Deploy Staging**: Automatically deploys to staging environment
  - SSH into staging server
  - Pull latest images
  - Run database migrations
  - Health checks
  - Smoke tests
- **Deploy Production**: Deploys to production (requires manual approval)
  - Creates database backup before deployment
  - Zero-downtime rolling restart
  - Health checks
  - Creates GitHub release on success

**Environments**:

- **Staging**: https://staging.volunteersync.example.com
- **Production**: https://volunteersync.example.com (manual approval required)

**Manual Triggers**:

- Deploy to specific environment (staging or production)
- Rollback to previous version

**Required Secrets**:

- `STAGING_SSH_PRIVATE_KEY`, `STAGING_SSH_HOST`, `STAGING_SSH_USER`
- `PRODUCTION_SSH_PRIVATE_KEY`, `PRODUCTION_SSH_HOST`, `PRODUCTION_SSH_USER`
- `NEXT_PUBLIC_API_URL_STAGING`, `NEXT_PUBLIC_API_URL_PRODUCTION`

**Deployment Flow**:

```
main branch → Build Images → Deploy Staging → (Manual Approval) → Deploy Production
```

---

### 3. Security Scanning (`security.yml`)

**Trigger**:

- Push to main, feature/_, bugfix/_
- Pull requests to main
- Daily at 2 AM UTC (scheduled)
- Manual workflow dispatch

**Purpose**: Automated security vulnerability scanning and dependency checks

**Jobs**:

- **Backend Dependency Scan**:
  - Gosec (Go security scanner)
  - Nancy (Go dependency vulnerability scanner)
  - govulncheck (official Go vulnerability scanner)
- **Frontend Dependency Scan**:
  - npm audit (Node.js dependency vulnerabilities)
  - Generates audit fix report
- **CodeQL Analysis**:
  - Semantic code analysis for Go and JavaScript/TypeScript
  - Security-extended and quality queries
  - Uploads results to GitHub Security tab
- **Docker Image Scan**:
  - Trivy scanner for backend and frontend Docker images
  - Scans for OS and application vulnerabilities
- **Secret Scanning**:
  - Gitleaks for detecting exposed secrets in code
- **OWASP Dependency-Check**:
  - Comprehensive dependency vulnerability analysis
  - Fails on CVSS score ≥ 7
- **Security Summary**: Aggregates results and creates summary

**Automated Actions**:

- Uploads SARIF reports to GitHub Security
- Creates GitHub issue for critical vulnerabilities (scheduled runs only)
- Fails workflow if critical vulnerabilities detected (CVSS ≥ 7)

**Artifacts**:

- npm audit report (30-day retention)
- OWASP Dependency-Check HTML report (30-day retention)

---

## Setup Instructions

### 1. Required Secrets

Configure the following secrets in GitHub repository settings (Settings → Secrets and variables → Actions):

**Staging Environment**:

```
STAGING_SSH_PRIVATE_KEY   # SSH private key for staging server
STAGING_SSH_HOST          # Staging server hostname or IP
STAGING_SSH_USER          # SSH username for staging server
NEXT_PUBLIC_API_URL_STAGING  # API URL for staging frontend
```

**Production Environment**:

```
PRODUCTION_SSH_PRIVATE_KEY   # SSH private key for production server
PRODUCTION_SSH_HOST          # Production server hostname or IP
PRODUCTION_SSH_USER          # SSH username for production server
NEXT_PUBLIC_API_URL_PRODUCTION  # API URL for production frontend
```

### 2. GitHub Container Registry (GHCR)

The workflows use GitHub Container Registry to store Docker images. No additional setup required - uses `GITHUB_TOKEN` automatically.

Images are pushed to:

- `ghcr.io/<owner>/volunteersync-backend`
- `ghcr.io/<owner>/volunteersync-frontend`

### 3. Codecov Integration (Optional)

For code coverage reporting, configure Codecov:

1. Sign up at https://codecov.io
2. Add repository
3. No token needed for public repos; for private repos, add `CODECOV_TOKEN` secret

### 4. Server Setup

**Staging and Production Servers** must have:

- Docker and Docker Compose installed
- SSH access configured
- Directory `/opt/volunteersync` with docker-compose.prod.yml
- User has sudo/docker permissions

**Server Directory Structure**:

```
/opt/volunteersync/
├── docker-compose.prod.yml
├── .env                     # Environment variables
└── backups/                 # Database backups
```

### 5. Environment Protection Rules

Configure environment protection in GitHub (Settings → Environments):

**Staging**:

- No approvals required
- No wait time

**Production**:

- Required reviewers: 1-2 team members
- Wait timer: Optional (e.g., 5 minutes)
- Branch protection: Only main branch can deploy

---

## Workflow Best Practices

### For Developers

1. **Before Pushing**:

   - Run `go fmt` and `go vet` for backend
   - Run `npm run lint` and `npm run format` for frontend
   - Run tests locally: `go test ./...` and `npm test`

2. **Pull Requests**:

   - CI pipeline must pass before merge
   - Review Codecov reports for test coverage changes
   - Check security scan results

3. **Merging to Main**:
   - Triggers automatic staging deployment
   - Monitor staging health checks
   - Production deployment requires manual approval

### For Maintainers

1. **Security Alerts**:

   - Review daily security scan reports
   - Address critical/high vulnerabilities within 7 days
   - Update dependencies monthly

2. **Production Deployments**:

   - Verify staging deployment is stable
   - Approve production deployment during low-traffic hours
   - Monitor health checks and rollback if needed

3. **Rollback Procedure**:
   - Use "Rollback Deployment" workflow (manual dispatch)
   - Or SSH into server and run: `docker compose -f docker-compose.prod.yml down && docker compose -f docker-compose.prod.yml up -d`

---

## Monitoring & Logs

### GitHub Actions Logs

- View workflow runs: Actions tab in repository
- Download artifacts for detailed reports
- Check GitHub Security tab for vulnerability reports

### Server Logs

```bash
# View backend logs
docker compose -f docker-compose.prod.yml logs -f backend

# View frontend logs
docker compose -f docker-compose.prod.yml logs -f frontend

# View all services
docker compose -f docker-compose.prod.yml logs -f
```

### Health Checks

- Staging: `https://staging.volunteersync.example.com/api/health`
- Production: `https://volunteersync.example.com/api/health`

---

## Troubleshooting

### CI Pipeline Fails

**Linting Errors**:

```bash
# Fix backend linting
cd backend && golangci-lint run --fix

# Fix frontend linting
cd frontend && npm run lint:fix && npm run format
```

**Test Failures**:

- Check test logs in workflow run
- Run tests locally with same environment
- Ensure database migrations are up to date

**Coverage Below 80%**:

- Add tests for uncovered code
- Check coverage report in artifacts

### CD Pipeline Fails

**SSH Connection Issues**:

- Verify SSH key is correct (no extra spaces/newlines)
- Check server is accessible: `ssh user@host`
- Ensure SSH key has correct permissions (600)

**Docker Image Pull Fails**:

- Check GHCR permissions
- Verify GITHUB_TOKEN has package write access

**Health Check Fails**:

- Check server logs: `docker compose logs`
- Verify migrations ran successfully
- Check database connectivity

### Security Scan Fails

**Critical Vulnerabilities Detected**:

- Review vulnerability details in Security tab
- Update affected dependencies
- Check if patches are available

**False Positives**:

- Review findings and dismiss if not applicable
- Add exceptions to scanner configuration if needed

---

## Performance

**Typical Execution Times**:

- CI Pipeline: 8-12 minutes
- CD Pipeline (Staging): 5-7 minutes
- CD Pipeline (Production): 7-10 minutes
- Security Scans: 10-15 minutes

**Optimization Tips**:

- Caching reduces build time by 30-50%
- Parallel jobs reduce overall pipeline time
- E2E tests can be run selectively for faster feedback

---

## Future Enhancements

- [ ] Add performance testing (Lighthouse CI)
- [ ] Add visual regression testing (Percy, Chromatic)
- [ ] Add Kubernetes deployment support
- [ ] Add canary deployments
- [ ] Add automated rollback on health check failure
- [ ] Add Slack/email notifications
- [ ] Add load testing (k6, Artillery)

---

## Support

For issues with CI/CD pipelines:

1. Check workflow logs in Actions tab
2. Review this README
3. Contact DevOps team
4. Create issue with `ci/cd` label
