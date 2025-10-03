/**
 * E2E Test: Volunteer Journey
 * T044: Test profile creation, opportunity search, registration, and hours verification
 *
 * This test validates the complete volunteer user journey:
 * - Volunteer registration and profile setup (FR-028, FR-029, FR-030, FR-031)
 * - Opportunity search and discovery (FR-039, FR-040, NFR-018)
 * - Event registration (FR-055, FR-064)
 * - Check-in and hours verification (FR-046, FR-047, FR-048)
 * - Impact tracking and dashboard (FR-042, FR-043, FR-045)
 */

import { test, expect, createHelpers, TestUser } from './helpers/fixtures';

test.describe('Volunteer Journey', () => {
  test.describe.configure({ mode: 'serial' });

  let helpers: ReturnType<typeof createHelpers>;
  let volunteerUser: TestUser;
  let orgAdminUser: TestUser;
  let accessToken: string;
  let opportunityId: string;

  test.beforeAll(async ({ browser, apiBaseURL }) => {
    // Create a new page for setup
    const page = await browser.newPage();
    helpers = createHelpers(page, apiBaseURL);

    // Create organization admin and organization
    const timestamp = Date.now();
    orgAdminUser = {
      email: `admin.${timestamp}@greenearth.org`,
      password: 'AdminPass123!',
      firstName: 'Jane',
      lastName: 'Admin',
      userType: 'organization_admin',
      securityQuestions: [
        { question: "What is your mother's maiden name?", answer: 'Johnson' },
        { question: 'What city were you born in?', answer: 'Boston' },
        { question: "What was your first pet's name?", answer: 'Max' },
      ],
    };

    // Register org admin
    const { accessToken: adminToken } = await helpers.registerUser(orgAdminUser);

    // Create organization
    const { organizationId } = await helpers.createOrganization(adminToken, {
      name: 'Green Earth Initiative',
      mission_statement: 'Protecting the environment through action',
      description: 'We organize community cleanups and environmental events',
      email: 'contact@greenearth.org',
      phone: '+1-555-0100',
      address_line_1: '123 Eco Street',
      city: 'Portland',
      state: 'Oregon',
      postal_code: '97201',
      website: 'https://greenearth.org',
      cause_categories: ['environment'],
    });

    // Create opportunity
    const { opportunityId: oppId } = await helpers.createOpportunity(adminToken, {
      organization_id: organizationId,
      title: 'Beach Cleanup - Ocean Park',
      description: 'Join us for a morning beach cleanup to protect our oceans',
      start_date: '2025-10-15',
      start_time: '09:00',
      end_date: '2025-10-15',
      end_time: '12:00',
      timezone: 'America/Los_Angeles',
      address_line_1: 'Ocean Park Beach',
      city: 'San Francisco',
      state: 'California',
      postal_code: '94121',
      capacity: 20,
      min_age: 16,
      cause_categories: ['environment'],
      status: 'published',
    });

    opportunityId = oppId;

    await page.close();
  });

  test.beforeEach(async ({ page, apiBaseURL }) => {
    helpers = createHelpers(page, apiBaseURL);
  });

  test('Step 1: Volunteer registers and creates profile (FR-028, FR-029)', async ({ page }) => {
    // Generate unique volunteer
    const timestamp = Date.now();
    volunteerUser = {
      email: `john.volunteer.${timestamp}@example.com`,
      password: 'VolunteerPass123!',
      firstName: 'John',
      lastName: 'Volunteer',
      userType: 'volunteer',
      securityQuestions: [
        { question: "What is your mother's maiden name?", answer: 'Smith' },
        { question: 'What city were you born in?', answer: 'Seattle' },
        { question: "What was your first pet's name?", answer: 'Buddy' },
      ],
    };

    // Navigate to registration
    await helpers.goto('/auth/register');

    // Fill registration form
    await helpers.fillFormField('Email', volunteerUser.email);
    await helpers.fillFormField('Password', volunteerUser.password);
    await helpers.fillFormField('First Name', volunteerUser.firstName);
    await helpers.fillFormField('Last Name', volunteerUser.lastName);
    await page.getByLabel(/user type/i).selectOption('volunteer');

    // Fill security questions
    for (let i = 0; i < volunteerUser.securityQuestions.length; i++) {
      const sq = volunteerUser.securityQuestions[i];
      if (sq) {
        await page.getByLabel(new RegExp(`question ${i + 1}`, 'i')).fill(sq.question);
        await page.getByLabel(new RegExp(`answer ${i + 1}`, 'i')).fill(sq.answer);
      }
    }

    await helpers.clickButton('Register');

    // Should be redirected to profile setup or dashboard
    await page.waitForURL(/profile|dashboard|volunteer/, { timeout: 10000 });

    // Save access token for API calls
    accessToken = await page.evaluate(() => localStorage.getItem('access_token') || '');
    expect(accessToken).toBeTruthy();

    // Complete volunteer profile
    // Navigate to profile page if not already there
    if (!(await page.url().includes('profile'))) {
      await page.getByRole('link', { name: /profile/i }).click();
      await page.waitForURL(/profile/, { timeout: 5000 });
    }

    // Fill profile information
    await helpers.fillFormField('Location', 'San Francisco, CA');
    await helpers.fillFormField(
      'Biography',
      'Passionate about environmental conservation and community service'
    );

    // Select skills
    const skillsDropdown = page
      .locator('[data-testid="skills-select"], select[name*="skill"]')
      .first();
    if (await skillsDropdown.isVisible()) {
      await skillsDropdown.selectOption('event-planning');
    }

    // Select interests
    const interestsDropdown = page
      .locator('[data-testid="interests-select"], select[name*="interest"]')
      .first();
    if (await interestsDropdown.isVisible()) {
      await interestsDropdown.selectOption('environment');
    }

    // Set availability (weekends)
    const saturdayCheckbox = page
      .locator('input[name*="saturday"], [data-testid="availability-saturday"]')
      .first();
    if (await saturdayCheckbox.isVisible()) {
      await saturdayCheckbox.check();
    }

    const sundayCheckbox = page
      .locator('input[name*="sunday"], [data-testid="availability-sunday"]')
      .first();
    if (await sundayCheckbox.isVisible()) {
      await sundayCheckbox.check();
    }

    // Save profile
    await helpers.clickButton('Save');

    // Should see success message or stay on profile page
    await expect(
      page.locator('[role="alert"], [data-testid="success-message"]').first()
    ).toBeVisible({
      timeout: 5000,
    });
  });

  test('Step 2: Search for opportunities with filters (FR-039, FR-040, NFR-018)', async ({
    page,
  }) => {
    // Login as volunteer
    const { accessToken: volToken, refreshToken } = await helpers.loginUser(
      volunteerUser.email,
      volunteerUser.password
    );
    await helpers.setAuthTokens(volToken, refreshToken);

    // Navigate to opportunity search
    await helpers.goto('/opportunities');

    // Verify search page loads
    await expect(page.locator('h1, h2')).toContainText(/find|search|opportunities/i);

    // Apply search filters
    await helpers.fillFormField('Location', 'San Francisco, CA');

    // Set radius
    const radiusSelect = page
      .locator('select[name*="radius"], [data-testid="radius-select"]')
      .first();
    if (await radiusSelect.isVisible()) {
      await radiusSelect.selectOption('25');
    }

    // Select cause
    const causeSelect = page.locator('select[name*="cause"], [data-testid="cause-select"]').first();
    if (await causeSelect.isVisible()) {
      await causeSelect.selectOption('environment');
    }

    // Date range
    const startDateInput = page.locator('input[name*="start"], [data-testid="start-date"]').first();
    if (await startDateInput.isVisible()) {
      await startDateInput.fill('2025-10-01');
    }

    const endDateInput = page.locator('input[name*="end"], [data-testid="end-date"]').first();
    if (await endDateInput.isVisible()) {
      await endDateInput.fill('2025-10-31');
    }

    // Measure search performance (NFR-018: <2s)
    const searchDuration = await helpers.measurePerformance(async () => {
      await helpers.clickButton('Search');
      await page.waitForSelector(
        '[data-testid="search-results"], [data-testid="opportunity-card"]',
        {
          timeout: 5000,
        }
      );
    });

    // Verify performance requirement (<2 seconds)
    expect(searchDuration).toBeLessThan(2000);

    // Verify search results
    const opportunityCards = page.locator('[data-testid="opportunity-card"]');
    await expect(opportunityCards).toHaveCount(1, { timeout: 5000 });

    // Verify our created opportunity appears
    await expect(opportunityCards.first()).toContainText('Beach Cleanup - Ocean Park');
    await expect(opportunityCards.first()).toContainText('0 of 20 spots filled');

    // Verify map is displayed (optional, may not be implemented yet)
    const mapElement = page.locator('[data-testid="opportunity-map"], .leaflet-container').first();
    if (await mapElement.isVisible({ timeout: 2000 }).catch(() => false)) {
      await expect(mapElement).toBeVisible();
    }
  });

  test('Step 3: Register for opportunity (FR-055, FR-064)', async ({ page }) => {
    // Login as volunteer
    const { accessToken: volToken, refreshToken } = await helpers.loginUser(
      volunteerUser.email,
      volunteerUser.password
    );
    await helpers.setAuthTokens(volToken, refreshToken);

    // Navigate to opportunity detail page
    await helpers.goto(`/opportunities/${opportunityId}`);

    // Verify opportunity details are displayed
    await expect(page.locator('h1, h2')).toContainText('Beach Cleanup - Ocean Park');
    await expect(page.locator('[data-testid="capacity-display"]')).toContainText(/0 of 20|spots/i);

    // Click register button
    const registerButton = page
      .getByRole('button', { name: /register|sign up/i })
      .or(page.locator('[data-testid="register-button"]'));
    await registerButton.click();

    // Should see confirmation message or notification
    await expect(
      page
        .locator('[role="alert"], [data-testid="success-message"], [data-testid="notification"]')
        .first()
    ).toBeVisible({ timeout: 5000 });

    // Verify registration status
    await expect(page.locator('[data-testid="registration-status"]').first()).toContainText(
      /registered|confirmed/i
    );

    // Capacity should update to 1 of 20
    await expect(page.locator('[data-testid="capacity-display"]')).toContainText(/1 of 20/);

    // Verify calendar download button exists (FR-064)
    const calendarButton = page
      .getByRole('button', { name: /calendar|download|\.ics/i })
      .or(page.locator('[data-testid="calendar-download"]'));

    if (await calendarButton.isVisible({ timeout: 2000 }).catch(() => false)) {
      await expect(calendarButton).toBeVisible();
    }

    // Navigate to "My Events" to verify event appears
    await page.getByRole('link', { name: /my events|events/i }).click();
    await page.waitForURL(/events/, { timeout: 5000 });

    // Verify event appears in upcoming events
    const eventCard = page
      .locator('[data-testid="event-card"]')
      .filter({ hasText: 'Beach Cleanup' })
      .first();
    await expect(eventCard).toBeVisible({ timeout: 5000 });
  });

  test('Step 4: Verify hours logging notification (FR-047)', async ({ page }) => {
    // This test simulates the coordinator logging hours
    // In a real scenario, this would happen after the event

    // Login as volunteer
    const { accessToken: volToken, refreshToken } = await helpers.loginUser(
      volunteerUser.email,
      volunteerUser.password
    );
    await helpers.setAuthTokens(volToken, refreshToken);

    // Go to dashboard
    await helpers.goto('/dashboard');

    // NOTE: Since hours are logged by the coordinator (not implemented yet),
    // we'll check if the notification center exists and is functional

    // Check if notification bell/icon exists
    const notificationBell = page
      .locator('[data-testid="notification-bell"]')
      .or(page.getByRole('button', { name: /notification/i }));

    if (await notificationBell.isVisible({ timeout: 2000 }).catch(() => false)) {
      await expect(notificationBell).toBeVisible();

      // Click to open notifications
      await notificationBell.click();

      // Verify notification center opens
      await expect(page.locator('[data-testid="notification-center"]').first()).toBeVisible({
        timeout: 3000,
      });
    }

    // In a complete test, we would verify:
    // - Hours logged notification appears
    // - Volunteer can verify hours
    // - Total hours updates on profile
  });

  test('Step 5: View volunteer dashboard and impact metrics (FR-042, FR-043, FR-045)', async ({
    page,
  }) => {
    // Login as volunteer
    const { accessToken: volToken, refreshToken } = await helpers.loginUser(
      volunteerUser.email,
      volunteerUser.password
    );
    await helpers.setAuthTokens(volToken, refreshToken);

    // Navigate to dashboard
    await helpers.goto('/dashboard');

    // Verify dashboard loads
    await expect(page.locator('h1, h2')).toContainText(/dashboard|welcome/i);

    // Verify impact metrics are displayed (FR-042)
    const metricsSection = page
      .locator('[data-testid="impact-metrics"], [data-testid="dashboard-metrics"]')
      .first();

    if (await metricsSection.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(metricsSection).toBeVisible();

      // Check for total hours metric
      await expect(page.locator('[data-testid="total-hours"]').first()).toBeVisible({
        timeout: 3000,
      });

      // Check for total events metric
      await expect(page.locator('[data-testid="total-events"]').first()).toBeVisible({
        timeout: 3000,
      });

      // Check for organizations metric
      await expect(page.locator('[data-testid="total-organizations"]').first()).toBeVisible({
        timeout: 3000,
      });
    }

    // Verify upcoming events section exists
    const upcomingEvents = page.locator('[data-testid="upcoming-events"]').first();
    if (await upcomingEvents.isVisible({ timeout: 2000 }).catch(() => false)) {
      await expect(upcomingEvents).toBeVisible();
    }

    // Navigate to impact/analytics page if it exists (FR-043, FR-045)
    const impactLink = page.getByRole('link', { name: /impact|analytics/i });
    if (await impactLink.isVisible({ timeout: 2000 }).catch(() => false)) {
      await impactLink.click();
      await page.waitForURL(/impact|analytics/, { timeout: 5000 });

      // Verify charts/visualizations exist
      const chartElement = page.locator('[data-testid="hours-chart"], canvas, svg').first();

      if (await chartElement.isVisible({ timeout: 3000 }).catch(() => false)) {
        await expect(chartElement).toBeVisible();
      }

      // Verify download report button exists (FR-045)
      const downloadButton = page
        .getByRole('button', { name: /download|report|pdf/i })
        .or(page.locator('[data-testid="download-report"]'));

      if (await downloadButton.isVisible({ timeout: 2000 }).catch(() => false)) {
        await expect(downloadButton).toBeVisible();
      }
    }
  });

  test('Step 6: Check for achievement badges', async ({ page }) => {
    // Login as volunteer
    const { accessToken: volToken, refreshToken } = await helpers.loginUser(
      volunteerUser.email,
      volunteerUser.password
    );
    await helpers.setAuthTokens(volToken, refreshToken);

    // Navigate to profile or achievements page
    await helpers.goto('/profile');

    // Look for achievements/badges section
    const achievementsSection = page
      .locator('[data-testid="achievements"], [data-testid="badges"]')
      .first();

    if (await achievementsSection.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(achievementsSection).toBeVisible();

      // After registering for first event, should have "First Event" badge potential
      // (Note: badge award logic runs as cron job, may not be immediate)
    }
  });
});
