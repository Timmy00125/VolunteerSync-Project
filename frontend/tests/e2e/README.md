# Frontend E2E Tests

This directory contains End-to-End (E2E) tests for the VolunteerSync platform using Playwright.

## Overview

The E2E tests validate complete user journeys through the application:

- **T043: Authentication Flow** (`auth.spec.ts`)
  - User registration with security questions (FR-002)
  - Login/logout functionality (FR-001)
  - Password reset using security questions (FR-003)
  - Rate limiting enforcement
  - Validation of password strength and duplicate emails

- **T044: Volunteer Journey** (`volunteer_journey.spec.ts`)
  - Volunteer registration and profile creation (FR-028, FR-029, FR-030, FR-031)
  - Opportunity search with filters (FR-039, FR-040)
  - Event registration (FR-055, FR-064)
  - Hours verification workflow (FR-046, FR-047, FR-048)
  - Dashboard and impact tracking (FR-042, FR-043, FR-045)

- **T045: Organization Journey** (`org_journey.spec.ts`)
  - Organization admin registration (FR-002)
  - Organization profile creation with auto-verification (FR-015, FR-017)
  - Opportunity creation and publishing (FR-021)
  - Roster management and hours logging (FR-046, FR-047, FR-054)
  - Organization analytics (FR-023, FR-024, FR-025)

## Prerequisites

1. **Backend API** must be running at `http://localhost:8080/api/v1`
2. **Frontend application** must be running at `http://localhost:3000`
3. **Database** must be clean before running tests (or use transaction rollback)
4. **Node.js** 20+ and **npm/bun** installed

## Installation

```bash
# Install Playwright browsers (first time only)
npx playwright install

# Or using bun
bun playwright install
```

## Running Tests

### Run all E2E tests

```bash
npm run test:e2e

# Or with bun
bun run test:e2e
```

### Run specific test file

```bash
npx playwright test tests/e2e/auth.spec.ts
npx playwright test tests/e2e/volunteer_journey.spec.ts
npx playwright test tests/e2e/org_journey.spec.ts
```

### Run in headed mode (see browser)

```bash
npx playwright test --headed
```

### Run in debug mode

```bash
npx playwright test --debug
```

### Run in UI mode (interactive)

```bash
npx playwright test --ui
```

## Environment Variables

Configure the following environment variables if needed:

```bash
# Base URL for frontend application
BASE_URL=http://localhost:3000

# API base URL for backend
API_BASE_URL=http://localhost:8080/api/v1
```

## Test Structure

```
tests/e2e/
├── helpers/
│   └── fixtures.ts          # Test fixtures, helpers, and utilities
├── auth.spec.ts             # T043: Authentication flow tests
├── volunteer_journey.spec.ts # T044: Volunteer journey tests
└── org_journey.spec.ts      # T045: Organization journey tests
```

## Test Fixtures

Custom fixtures provide:

- `apiBaseURL`: API base URL
- `volunteerUser`: Pre-configured volunteer test user
- `orgAdminUser`: Pre-configured org admin test user
- `TestHelpers`: Utility functions for common actions

### Using Fixtures

```typescript
import { test, expect, createHelpers } from './helpers/fixtures';

test('example test', async ({ page, apiBaseURL, volunteerUser }) => {
  const helpers = createHelpers(page, apiBaseURL);

  // Use helpers
  await helpers.registerUser(volunteerUser);
  await helpers.goto('/dashboard');

  // Use Playwright APIs
  await expect(page).toHaveURL(/dashboard/);
});
```

## Helper Functions

The `TestHelpers` class provides:

- `registerUser(user)`: Register user via API
- `loginUser(email, password)`: Login via API
- `setAuthTokens(access, refresh)`: Set tokens in browser storage
- `clearAuthTokens()`: Clear auth tokens
- `fillFormField(label, value)`: Fill form input
- `clickButton(text)`: Click button by text
- `waitForNotification(message?)`: Wait for toast/alert
- `goto(path)`: Navigate with wait for load
- `isLoggedIn()`: Check if user has valid token
- `createOrganization(token, data)`: Create org via API
- `createOpportunity(token, data)`: Create opportunity via API
- `measurePerformance(action)`: Measure action duration

## Performance Requirements

The tests validate these performance requirements:

- **NFR-018**: Opportunity search results in <2 seconds
- **NFR-017**: Organization creation completes in <5 minutes
- **NFR-006**: Event registration completes in <1 second (where tested)

## Reports

After running tests, view the HTML report:

```bash
npx playwright show-report
```

## Debugging

### Take screenshots

Screenshots are automatically taken on test failure and saved to `test-results/`.

### Record video

Videos are recorded on failure and saved to `test-results/`.

### VS Code Extension

Install the Playwright VS Code extension for better debugging:

```bash
code --install-extension ms-playwright.playwright
```

## Best Practices

1. **Test Isolation**: Each test should be independent and not rely on other tests
2. **Clean State**: Use `beforeEach` to ensure clean state
3. **Explicit Waits**: Use `waitFor` or `expect` with timeouts instead of hard waits
4. **Data Attributes**: Prefer `data-testid` selectors over CSS classes
5. **Error Handling**: Tests should handle gracefully when UI elements aren't implemented yet

## CI/CD Integration

Tests are configured to run in GitHub Actions:

```yaml
- name: Run E2E tests
  run: |
    npx playwright install --with-deps
    npm run test:e2e
```

## Troubleshooting

### Tests fail with "Element not found"

- Check if the frontend UI is fully implemented
- Verify data-testid attributes match test selectors
- Use headed mode (`--headed`) to see what's happening

### Authentication tests fail

- Ensure backend API is running
- Check API base URL is correct
- Verify database is accessible

### Performance tests fail

- Increase timeout if running on slower hardware
- Check backend API response times
- Verify database queries are optimized

## Next Steps

After implementing frontend UI and backend API:

1. Run tests and fix any failures
2. Add more test scenarios for edge cases
3. Configure tests to run in CI/CD pipeline
4. Add visual regression testing with Playwright screenshots

## Resources

- [Playwright Documentation](https://playwright.dev)
- [Best Practices](https://playwright.dev/docs/best-practices)
- [API Reference](https://playwright.dev/docs/api/class-playwright)
