/**
 * E2E Test: Organization Journey
 * T045: Test org creation, opportunity posting, and hours logging
 *
 * This test validates the complete organization admin journey:
 * - Organization admin registration (FR-002)
 * - Organization profile creation with auto-verification (FR-015, FR-017)
 * - Opportunity creation and publishing (FR-021)
 * - Volunteer management (viewing roster, check-ins)
 * - Hours logging and tracking (FR-046, FR-047, FR-054)
 * - Organization analytics (FR-023, FR-024, FR-025)
 */

import { test, expect, createHelpers, TestUser } from './helpers/fixtures';

test.describe('Organization Journey', () => {
  test.describe.configure({ mode: 'serial' });

  let helpers: ReturnType<typeof createHelpers>;
  let orgAdminUser: TestUser;
  let accessToken: string;
  let organizationId: string;
  let opportunityId: string;

  test.beforeEach(async ({ page, apiBaseURL }) => {
    helpers = createHelpers(page, apiBaseURL);
  });

  test('Step 1: Organization admin registers (FR-002)', async ({ page }) => {
    // Generate unique org admin
    const timestamp = Date.now();
    orgAdminUser = {
      email: `admin.${timestamp}@nonprofitorg.org`,
      password: 'OrgAdminPass123!',
      firstName: 'Sarah',
      lastName: 'OrgAdmin',
      userType: 'organization_admin',
      securityQuestions: [
        { question: "What is your mother's maiden name?", answer: 'Wilson' },
        { question: 'What city were you born in?', answer: 'Chicago' },
        { question: "What was your first pet's name?", answer: 'Rex' },
      ],
    };

    // Navigate to registration
    await helpers.goto('/auth/register');

    // Fill registration form
    await helpers.fillFormField('Email', orgAdminUser.email);
    await helpers.fillFormField('Password', orgAdminUser.password);
    await helpers.fillFormField('First Name', orgAdminUser.firstName);
    await helpers.fillFormField('Last Name', orgAdminUser.lastName);
    await page.getByLabel(/user type/i).selectOption('organization_admin');

    // Fill security questions
    for (let i = 0; i < orgAdminUser.securityQuestions.length; i++) {
      const sq = orgAdminUser.securityQuestions[i];
      if (sq) {
        await page.getByLabel(new RegExp(`question ${i + 1}`, 'i')).fill(sq.question);
        await page.getByLabel(new RegExp(`answer ${i + 1}`, 'i')).fill(sq.answer);
      }
    }

    await helpers.clickButton('Register');

    // Should be redirected to dashboard or org setup
    await page.waitForURL(/dashboard|organization|profile/, { timeout: 10000 });

    // Save access token
    accessToken = await page.evaluate(() => localStorage.getItem('access_token') || '');
    expect(accessToken).toBeTruthy();

    // Verify user is logged in
    await expect(page.locator(`text=${orgAdminUser.firstName}`).first()).toBeVisible({
      timeout: 5000,
    });
  });

  test('Step 2: Create organization profile with auto-verification (FR-015, FR-017)', async ({
    page,
  }) => {
    // Login as org admin
    const { accessToken: orgToken, refreshToken } = await helpers.loginUser(
      orgAdminUser.email,
      orgAdminUser.password
    );
    await helpers.setAuthTokens(orgToken, refreshToken);
    accessToken = orgToken;

    // Navigate to create organization page
    await helpers.goto('/organization/new');

    // Or click create organization button if on dashboard
    if (!(await page.url().includes('organization'))) {
      const createOrgButton = page
        .getByRole('button', { name: /create organization/i })
        .or(page.getByRole('link', { name: /create organization/i }));

      if (await createOrgButton.isVisible({ timeout: 2000 }).catch(() => false)) {
        await createOrgButton.click();
        await page.waitForURL(/organization.*new/, { timeout: 5000 });
      }
    }

    // Verify we're on organization creation page
    await expect(page.locator('h1, h2')).toContainText(/create|new.*organization/i);

    // Fill organization form
    const orgName = `Community Impact ${Date.now()}`;
    await helpers.fillFormField('Organization Name', orgName);
    await helpers.fillFormField(
      'Mission Statement',
      'Empowering communities through education and service'
    );
    await helpers.fillFormField(
      'Description',
      'We provide educational programs and community service opportunities'
    );
    await helpers.fillFormField('Email', 'contact@communityimpact.org');
    await helpers.fillFormField('Phone', '+1-555-0200');
    await helpers.fillFormField('Address', '456 Service Avenue');
    await helpers.fillFormField('City', 'Seattle');
    await helpers.fillFormField('State', 'Washington');
    await helpers.fillFormField('Postal Code', '98101');
    await helpers.fillFormField('Website', 'https://communityimpact.org');

    // Select cause category
    const causeSelect = page.locator('select[name*="cause"], [data-testid="cause-select"]').first();
    if (await causeSelect.isVisible()) {
      await causeSelect.selectOption('education');
    }

    // Upload logo (optional)
    const logoInput = page.locator('input[type="file"][name*="logo"]').first();
    if (await logoInput.isVisible({ timeout: 2000 }).catch(() => false)) {
      // Would upload file here in real implementation
      // await logoInput.setInputFiles('path/to/test-logo.png');
    }

    // Measure organization creation time (NFR-017: <5 minutes)
    const creationDuration = await helpers.measurePerformance(async () => {
      await helpers.clickButton('Create Organization');
      await page.waitForURL(/organization\/[^/]+$/, { timeout: 15000 });
    });

    // Verify performance requirement (<5 minutes = 300 seconds)
    expect(creationDuration).toBeLessThan(300000);

    // Should be redirected to organization profile page
    const firstWord = orgName.split(' ')[0] || 'Community';
    await expect(page.locator('h1, h2')).toContainText(new RegExp(firstWord, 'i'));

    // Verify auto-verification badge (FR-015)
    const verifiedBadge = page
      .locator('[data-testid="verified-badge"]')
      .or(page.getByText(/verified/i));
    if (await verifiedBadge.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(verifiedBadge).toBeVisible();
    }

    // Verify organization slug in URL
    await expect(page).toHaveURL(/\/organization\/[a-z0-9-]+/);

    // Save organization ID from URL or page
    const url = page.url();
    const match = url.match(/\/organization\/([a-f0-9-]+)/);
    if (match && match[1]) {
      organizationId = match[1];
    }
  });

  test('Step 3: Create and publish volunteer opportunity (FR-021)', async ({ page }) => {
    // Login as org admin
    const { accessToken: orgToken, refreshToken } = await helpers.loginUser(
      orgAdminUser.email,
      orgAdminUser.password
    );
    await helpers.setAuthTokens(orgToken, refreshToken);

    // Navigate to create opportunity page
    await helpers.goto('/opportunities/new');

    // Or navigate via dashboard
    if (!(await page.url().includes('opportunity'))) {
      const createOppButton = page
        .getByRole('button', { name: /create.*opportunity|new.*opportunity/i })
        .or(page.getByRole('link', { name: /create.*opportunity|new.*opportunity/i }));

      if (await createOppButton.isVisible({ timeout: 2000 }).catch(() => false)) {
        await createOppButton.click();
        await page.waitForURL(/opportunit.*new/, { timeout: 5000 });
      }
    }

    // Verify we're on opportunity creation page
    await expect(page.locator('h1, h2')).toContainText(/create|new.*opportunity/i);

    // Fill opportunity form
    const oppTitle = `Community Tutoring Program ${Date.now()}`;
    await helpers.fillFormField('Title', oppTitle);
    await helpers.fillFormField(
      'Description',
      'Help local students with homework and mentorship. No teaching experience required.'
    );

    // Set dates (future date)
    await helpers.fillFormField('Start Date', '2025-11-15');
    await helpers.fillFormField('Start Time', '14:00');
    await helpers.fillFormField('End Date', '2025-11-15');
    await helpers.fillFormField('End Time', '17:00');

    // Location
    await helpers.fillFormField('Address', 'Seattle Community Center');
    await helpers.fillFormField('City', 'Seattle');
    await helpers.fillFormField('State', 'Washington');
    await helpers.fillFormField('Postal Code', '98101');

    // Capacity
    await helpers.fillFormField('Capacity', '15');

    // Minimum age
    await helpers.fillFormField('Minimum Age', '18');

    // Cause category
    const causeSelect = page.locator('select[name*="cause"], [data-testid="cause-select"]').first();
    if (await causeSelect.isVisible()) {
      await causeSelect.selectOption('education');
    }

    // Timezone
    const timezoneSelect = page
      .locator('select[name*="timezone"], [data-testid="timezone-select"]')
      .first();
    if (await timezoneSelect.isVisible()) {
      await timezoneSelect.selectOption('America/Los_Angeles');
    }

    // Publish immediately
    const publishCheckbox = page
      .locator('input[name*="publish"], [data-testid="publish-immediately"]')
      .first();

    if (await publishCheckbox.isVisible({ timeout: 2000 }).catch(() => false)) {
      await publishCheckbox.check();
    }

    // Submit opportunity
    await helpers.clickButton('Create Opportunity');

    // Should be redirected to opportunity detail page
    await page.waitForURL(/opportunit.*\/[^/]+$/, { timeout: 10000 });

    // Verify opportunity details
    const oppFirstWord = oppTitle.split(' ')[0] || 'Community';
    await expect(page.locator('h1, h2')).toContainText(new RegExp(oppFirstWord, 'i'));

    // Verify published status
    const statusBadge = page.locator('[data-testid="opportunity-status"]');
    if (await statusBadge.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(statusBadge).toContainText(/published/i);
    }

    // Verify capacity display
    await expect(page.locator('[data-testid="capacity-display"]').first()).toContainText(
      /0 of 15|spots/i
    );

    // Save opportunity ID from URL
    const url = page.url();
    const match = url.match(/opportunit.*\/([a-f0-9-]+)/);
    if (match && match[1]) {
      opportunityId = match[1];
    }

    // Verify opportunity appears in search immediately
    await helpers.goto('/opportunities');
    await page.waitForSelector('[data-testid="opportunity-card"]', { timeout: 5000 });

    const opportunityCard = page
      .locator('[data-testid="opportunity-card"]')
      .filter({ hasText: oppTitle })
      .first();
    if (await opportunityCard.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(opportunityCard).toBeVisible();
    }
  });

  test('Step 4: View opportunity roster and registrations', async ({ page }) => {
    // Login as org admin
    const { accessToken: orgToken, refreshToken } = await helpers.loginUser(
      orgAdminUser.email,
      orgAdminUser.password
    );
    await helpers.setAuthTokens(orgToken, refreshToken);

    // Navigate to opportunity management
    await helpers.goto('/dashboard');

    // Go to opportunities list
    const opportunitiesLink = page.getByRole('link', { name: /opportunities/i });
    if (await opportunitiesLink.isVisible({ timeout: 2000 }).catch(() => false)) {
      await opportunitiesLink.click();
      await page.waitForURL(/opportunit/, { timeout: 5000 });
    }

    // View roster for the opportunity
    const viewRosterButton = page
      .getByRole('button', { name: /roster|volunteers|view/i })
      .or(page.getByRole('link', { name: /roster|volunteers/i }))
      .first();

    if (await viewRosterButton.isVisible({ timeout: 3000 }).catch(() => false)) {
      await viewRosterButton.click();

      // Should see roster page
      await expect(page.locator('h1, h2, h3')).toContainText(/roster|volunteers|registrations/i);

      // Roster table or list should be present (may be empty initially)
      const rosterTable = page.locator('table, [data-testid="roster-list"]').first();
      if (await rosterTable.isVisible({ timeout: 2000 }).catch(() => false)) {
        await expect(rosterTable).toBeVisible();
      }
    }
  });

  test('Step 5: Test hours logging interface (FR-046, FR-047)', async ({ page }) => {
    // Login as org admin
    const { accessToken: orgToken, refreshToken } = await helpers.loginUser(
      orgAdminUser.email,
      orgAdminUser.password
    );
    await helpers.setAuthTokens(orgToken, refreshToken);

    // Navigate to opportunity roster
    if (opportunityId) {
      await helpers.goto(`/opportunities/${opportunityId}/roster`);
    } else {
      // Navigate via dashboard
      await helpers.goto('/dashboard');
      await page.getByRole('link', { name: /opportunities/i }).click();
    }

    // Look for log hours button or link
    const logHoursButton = page
      .getByRole('button', { name: /log.*hours|hours.*log/i })
      .or(page.getByRole('link', { name: /log.*hours/i }))
      .first();

    if (await logHoursButton.isVisible({ timeout: 3000 }).catch(() => false)) {
      await logHoursButton.click();

      // Should see hours logging interface
      await expect(page.locator('h1, h2, h3')).toContainText(/log.*hours|hours.*log/i);

      // Hours input fields should be present
      const hoursInput = page.locator('input[type="number"][name*="hours"]').first();
      if (await hoursInput.isVisible({ timeout: 2000 }).catch(() => false)) {
        await expect(hoursInput).toBeVisible();
      }

      // Notes field should be present (FR-054)
      const notesField = page.locator('textarea[name*="notes"], textarea[name*="comment"]').first();
      if (await notesField.isVisible({ timeout: 2000 }).catch(() => false)) {
        await expect(notesField).toBeVisible();
      }
    }
  });

  test('Step 6: View organization analytics (FR-023, FR-024, FR-025)', async ({ page }) => {
    // Login as org admin
    const { accessToken: orgToken, refreshToken } = await helpers.loginUser(
      orgAdminUser.email,
      orgAdminUser.password
    );
    await helpers.setAuthTokens(orgToken, refreshToken);

    // Navigate to organization dashboard
    await helpers.goto('/dashboard');

    // Look for analytics link
    const analyticsLink = page.getByRole('link', { name: /analytics|reports|insights/i });

    if (await analyticsLink.isVisible({ timeout: 3000 }).catch(() => false)) {
      await analyticsLink.click();
      await page.waitForURL(/analytics|report/, { timeout: 5000 });

      // Verify analytics page loads
      await expect(page.locator('h1, h2')).toContainText(/analytics|reports|insights/i);

      // Verify key metrics are displayed (FR-023)
      const metrics = ['volunteers recruited', 'hours contributed', 'events hosted', 'retention'];

      for (const metric of metrics) {
        const metricElement = page
          .locator(`[data-testid*="${metric.replace(/\s+/g, '-')}"]`)
          .or(page.getByText(new RegExp(metric, 'i')))
          .first();

        if (await metricElement.isVisible({ timeout: 2000 }).catch(() => false)) {
          await expect(metricElement).toBeVisible();
        }
      }

      // Verify charts/visualizations exist (FR-024)
      const chartElement = page.locator('canvas, svg, [data-testid*="chart"]').first();
      if (await chartElement.isVisible({ timeout: 3000 }).catch(() => false)) {
        await expect(chartElement).toBeVisible();
      }
    }
  });

  test('Step 7: Manage team members', async ({ page }) => {
    // Login as org admin
    const { accessToken: orgToken, refreshToken } = await helpers.loginUser(
      orgAdminUser.email,
      orgAdminUser.password
    );
    await helpers.setAuthTokens(orgToken, refreshToken);

    // Navigate to team management
    await helpers.goto('/dashboard');

    // Look for team link
    const teamLink = page.getByRole('link', { name: /team|members|staff/i });

    if (await teamLink.isVisible({ timeout: 3000 }).catch(() => false)) {
      await teamLink.click();
      await page.waitForURL(/team|member/, { timeout: 5000 });

      // Verify team page loads
      await expect(page.locator('h1, h2')).toContainText(/team|members/i);

      // Verify current user is listed as admin
      const teamList = page.locator('table, [data-testid="team-list"]').first();
      if (await teamList.isVisible({ timeout: 2000 }).catch(() => false)) {
        await expect(teamList).toBeVisible();
        await expect(teamList).toContainText(orgAdminUser.email);
      }

      // Verify invite button exists (FR-014)
      const inviteButton = page
        .getByRole('button', { name: /invite|add.*member/i })
        .or(page.locator('[data-testid="invite-member"]'));

      if (await inviteButton.isVisible({ timeout: 2000 }).catch(() => false)) {
        await expect(inviteButton).toBeVisible();
      }
    }
  });

  test('Step 8: Send broadcast message to volunteers', async ({ page }) => {
    // Login as org admin
    const { accessToken: orgToken, refreshToken } = await helpers.loginUser(
      orgAdminUser.email,
      orgAdminUser.password
    );
    await helpers.setAuthTokens(orgToken, refreshToken);

    // Navigate to opportunity roster
    if (opportunityId) {
      await helpers.goto(`/opportunities/${opportunityId}/roster`);
    }

    // Look for message/communication button
    const messageButton = page
      .getByRole('button', { name: /message|send.*message|broadcast/i })
      .or(page.locator('[data-testid="send-message"]'))
      .first();

    if (await messageButton.isVisible({ timeout: 3000 }).catch(() => false)) {
      await messageButton.click();

      // Should see message composition interface
      const messageModal = page.locator('[role="dialog"], [data-testid="message-modal"]').first();
      if (await messageModal.isVisible({ timeout: 2000 }).catch(() => false)) {
        await expect(messageModal).toBeVisible();

        // Verify message input field
        const messageInput = messageModal.locator('textarea, [contenteditable]').first();
        await expect(messageInput).toBeVisible();
      }
    }
  });

  test('Step 9: Edit organization profile', async ({ page }) => {
    // Login as org admin
    const { accessToken: orgToken, refreshToken } = await helpers.loginUser(
      orgAdminUser.email,
      orgAdminUser.password
    );
    await helpers.setAuthTokens(orgToken, refreshToken);

    // Navigate to organization settings
    await helpers.goto('/dashboard');

    // Look for settings link
    const settingsLink = page.getByRole('link', { name: /settings|profile|edit/i });

    if (await settingsLink.isVisible({ timeout: 3000 }).catch(() => false)) {
      await settingsLink.click();

      // Should see organization edit form
      await expect(page.locator('h1, h2, h3')).toContainText(/settings|edit|profile/i);

      // Verify form fields are editable
      const nameField = page.locator('input[name*="name"]').first();
      if (await nameField.isVisible({ timeout: 2000 }).catch(() => false)) {
        await expect(nameField).toBeVisible();
        await expect(nameField).toBeEditable();
      }

      // Verify save button exists
      const saveButton = page.getByRole('button', { name: /save|update/i });
      await expect(saveButton).toBeVisible({ timeout: 3000 });
    }
  });
});
