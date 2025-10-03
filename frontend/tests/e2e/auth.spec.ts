/**
 * E2E Test: Authentication Flow
 * T043: Test registration, login, logout, and password reset with security questions
 *
 * This test validates:
 * - User registration with security questions (FR-002)
 * - User login with rate limiting (FR-001)
 * - User logout
 * - Password reset flow using security questions (FR-003)
 */

import { test, expect, createHelpers, TestUser } from './helpers/fixtures';

test.describe('Authentication Flow', () => {
  test.describe.configure({ mode: 'serial' });

  let helpers: ReturnType<typeof createHelpers>;
  let testUser: TestUser;

  test.beforeEach(async ({ page, apiBaseURL }) => {
    helpers = createHelpers(page, apiBaseURL);

    // Generate unique user for this test
    const timestamp = Date.now();
    testUser = {
      email: `test.user.${timestamp}@example.com`,
      password: 'TestPassword123!',
      firstName: 'Test',
      lastName: 'User',
      userType: 'volunteer',
      securityQuestions: [
        { question: "What is your mother's maiden name?", answer: 'SecurityAnswer1' },
        { question: 'What city were you born in?', answer: 'TestCity' },
        { question: "What was your first pet's name?", answer: 'Fluffy' },
      ],
    };
  });

  test('should register a new user with immediate activation (FR-002)', async ({ page }) => {
    // Navigate to registration page
    await helpers.goto('/auth/register');

    // Verify we're on the registration page
    await expect(page.locator('h1, h2')).toContainText(/register|sign up/i);

    // Fill in registration form
    await helpers.fillFormField('Email', testUser.email);
    await helpers.fillFormField('Password', testUser.password);
    await helpers.fillFormField('First Name', testUser.firstName);
    await helpers.fillFormField('Last Name', testUser.lastName);

    // Select user type
    await page.getByLabel(/user type|account type/i).selectOption(testUser.userType);

    // Fill in security questions
    for (let i = 0; i < testUser.securityQuestions.length; i++) {
      const sq = testUser.securityQuestions[i];
      if (sq) {
        await page
          .getByLabel(new RegExp(`security question ${i + 1}|question ${i + 1}`, 'i'))
          .fill(sq.question);
        await page.getByLabel(new RegExp(`answer ${i + 1}`, 'i')).fill(sq.answer);
      }
    }

    // Submit registration form
    await helpers.clickButton('Register');

    // Should be redirected to dashboard after successful registration
    await page.waitForURL(/\/(dashboard|volunteer|profile)/, { timeout: 10000 });

    // Verify user is logged in - should have access token
    const isLoggedIn = await helpers.isLoggedIn();
    expect(isLoggedIn).toBe(true);

    // Verify welcome message or user name appears
    const welcomeElement = page
      .locator('[data-testid="welcome-message"], [data-testid="user-name"]')
      .first();
    await expect(welcomeElement).toBeVisible({ timeout: 5000 });
    await expect(welcomeElement).toContainText(new RegExp(testUser.firstName, 'i'));

    // Account should be immediately active (no verification needed)
    // This is verified by being able to access protected routes
    await expect(page).toHaveURL(/dashboard|volunteer|profile/);
  });

  test('should login with valid credentials (FR-001)', async ({ page }) => {
    // First, register the user via API
    await helpers.registerUser(testUser);

    // Clear any existing auth tokens
    await helpers.clearAuthTokens();

    // Navigate to login page
    await helpers.goto('/auth/login');

    // Verify we're on the login page
    await expect(page.locator('h1, h2')).toContainText(/log in|sign in/i);

    // Fill in login form
    await helpers.fillFormField('Email', testUser.email);
    await helpers.fillFormField('Password', testUser.password);

    // Submit login form
    await helpers.clickButton('Log In');

    // Should be redirected to dashboard
    await page.waitForURL(/\/(dashboard|volunteer)/, { timeout: 10000 });

    // Verify logged in
    const isLoggedIn = await helpers.isLoggedIn();
    expect(isLoggedIn).toBe(true);

    // Verify user info is displayed
    await expect(page.locator(`text=${testUser.firstName}`).first()).toBeVisible({ timeout: 5000 });
  });

  test('should reject login with invalid credentials', async ({ page }) => {
    // Navigate to login page
    await helpers.goto('/auth/login');

    // Fill in login form with wrong password
    await helpers.fillFormField('Email', testUser.email);
    await helpers.fillFormField('Password', 'WrongPassword123!');

    // Submit login form
    await helpers.clickButton('Log In');

    // Should see error message
    const errorMessage = page
      .locator('[role="alert"], [data-testid="error-message"], .error')
      .first();
    await expect(errorMessage).toBeVisible({ timeout: 5000 });
    await expect(errorMessage).toContainText(/invalid|incorrect|wrong/i);

    // Should NOT be logged in
    const isLoggedIn = await helpers.isLoggedIn();
    expect(isLoggedIn).toBe(false);

    // Should still be on login page
    await expect(page).toHaveURL(/login/);
  });

  test('should logout successfully', async ({ page }) => {
    // Register and login via API
    const { accessToken, refreshToken } = await helpers.registerUser(testUser);
    await helpers.goto('/');
    await helpers.setAuthTokens(accessToken, refreshToken);

    // Navigate to dashboard
    await helpers.goto('/dashboard');

    // Click logout button
    const logoutButton = page
      .locator('[data-testid="logout-button"]')
      .or(page.getByRole('button', { name: /log out|sign out/i }))
      .or(page.getByText(/log out|sign out/i));

    await logoutButton.click();

    // Should be redirected to login or home page
    await page.waitForURL(/\/(login|auth|$)/, { timeout: 10000 });

    // Verify logged out - tokens should be cleared
    const isLoggedIn = await helpers.isLoggedIn();
    expect(isLoggedIn).toBe(false);

    // Attempting to access protected route should redirect to login
    await page.goto('/dashboard');
    await page.waitForURL(/\/(login|auth)/, { timeout: 10000 });
  });

  test('should reset password using security questions (FR-003)', async ({ page }) => {
    // Register user first
    await helpers.registerUser(testUser);

    // Logout
    await helpers.clearAuthTokens();

    // Navigate to password reset page
    await helpers.goto('/auth/reset-password');

    // Step 1: Request password reset
    await helpers.fillFormField('Email', testUser.email);
    await helpers.clickButton('Continue');

    // Step 2: Answer security questions (2 of 3 correct)
    await page.waitForSelector('text=/security question/i', { timeout: 5000 });

    // Answer the first two security questions correctly
    for (let i = 0; i < 2; i++) {
      const question = testUser.securityQuestions[i];
      if (question) {
        const answerField = page
          .locator(`input[placeholder*="Answer"], input[name*="answer"]`)
          .nth(i);
        await answerField.fill(question.answer);
      }
    }

    // Submit security question answers
    await helpers.clickButton('Verify');

    // Step 3: Set new password
    await page.waitForSelector('text=/new password/i', { timeout: 5000 });

    const newPassword = 'NewPassword456!';
    await helpers.fillFormField('New Password', newPassword);
    await helpers.fillFormField('Confirm Password', newPassword);

    // Submit new password
    await helpers.clickButton('Reset Password');

    // Should be redirected to login or see success message
    await page.waitForURL(/login|success/, { timeout: 10000 });

    // Try logging in with new password
    await helpers.goto('/auth/login');
    await helpers.fillFormField('Email', testUser.email);
    await helpers.fillFormField('Password', newPassword);
    await helpers.clickButton('Log In');

    // Should successfully login with new password
    await page.waitForURL(/dashboard/, { timeout: 10000 });
    const isLoggedIn = await helpers.isLoggedIn();
    expect(isLoggedIn).toBe(true);
  });

  test('should enforce rate limiting on login attempts', async ({ page }) => {
    // This test attempts to login 6 times with wrong password
    // The 6th attempt should be rate limited (FR-001: 5 attempts per 15 min)

    await helpers.goto('/auth/login');

    // Attempt login 5 times with wrong password
    for (let i = 0; i < 5; i++) {
      await helpers.fillFormField('Email', `ratelimit.test.${Date.now()}@example.com`);
      await helpers.fillFormField('Password', 'WrongPassword123!');
      await helpers.clickButton('Log In');

      // Wait for error
      await page.waitForSelector('[role="alert"], .error', { timeout: 3000 });

      // Small delay between attempts
      await page.waitForTimeout(100);
    }

    // 6th attempt should be rate limited
    await helpers.fillFormField('Email', `ratelimit.test.${Date.now()}@example.com`);
    await helpers.fillFormField('Password', 'WrongPassword123!');
    await helpers.clickButton('Log In');

    // Should see rate limit error (429 status)
    const errorMessage = page.locator('[role="alert"], [data-testid="error-message"]').first();
    await expect(errorMessage).toBeVisible({ timeout: 5000 });
    await expect(errorMessage).toContainText(/rate limit|too many|try again/i);
  });

  test('should require strong password on registration', async ({ page }) => {
    await helpers.goto('/auth/register');

    // Try registering with weak password
    await helpers.fillFormField('Email', `weak.${Date.now()}@example.com`);
    await helpers.fillFormField('Password', 'weak'); // Too short, no numbers
    await helpers.fillFormField('First Name', 'Test');
    await helpers.fillFormField('Last Name', 'User');

    // Submit form
    await helpers.clickButton('Register');

    // Should see password strength error
    const errorMessage = page
      .locator('[role="alert"], .error, [data-testid="password-error"]')
      .first();
    await expect(errorMessage).toBeVisible({ timeout: 5000 });
    await expect(errorMessage).toContainText(/password|strength|minimum|characters/i);

    // Should NOT be registered
    await expect(page).toHaveURL(/register/);
  });

  test('should prevent duplicate email registration', async ({ page }) => {
    // Register user first
    await helpers.registerUser(testUser);

    // Try to register again with same email
    await helpers.goto('/auth/register');

    await helpers.fillFormField('Email', testUser.email);
    await helpers.fillFormField('Password', 'AnotherPassword123!');
    await helpers.fillFormField('First Name', 'Another');
    await helpers.fillFormField('Last Name', 'User');
    await page.getByLabel(/user type/i).selectOption('volunteer');

    // Fill security questions
    for (let i = 0; i < 3; i++) {
      await page.getByLabel(new RegExp(`question ${i + 1}`, 'i')).fill(`Question ${i + 1}`);
      await page.getByLabel(new RegExp(`answer ${i + 1}`, 'i')).fill(`Answer ${i + 1}`);
    }

    await helpers.clickButton('Register');

    // Should see duplicate email error (409 status)
    const errorMessage = page.locator('[role="alert"], [data-testid="error-message"]').first();
    await expect(errorMessage).toBeVisible({ timeout: 5000 });
    await expect(errorMessage).toContainText(/already exists|already registered|duplicate/i);
  });
});
