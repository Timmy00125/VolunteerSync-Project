import { test as base, expect as baseExpect, Page } from '@playwright/test';

/**
 * Test Fixtures for VolunteerSync E2E Tests
 * Provides common utilities, authentication helpers, and API clients
 */

export interface TestUser {
  email: string;
  password: string;
  firstName: string;
  lastName: string;
  userType: 'volunteer' | 'organization_admin';
  securityQuestions: Array<{
    question: string;
    answer: string;
  }>;
}

export interface ExtendedFixtures {
  apiBaseURL: string;
  authenticatedPage: Page;
  volunteerUser: TestUser;
  orgAdminUser: TestUser;
}

/**
 * Extended test with custom fixtures
 */
export const test = base.extend<ExtendedFixtures>({
  apiBaseURL: async ({}, use: (value: string) => Promise<void>) => {
    const url = process.env.API_BASE_URL || 'http://localhost:8080/api/v1';
    await use(url);
  },

  volunteerUser: async ({}, use: (value: TestUser) => Promise<void>) => {
    const timestamp = Date.now();
    const user: TestUser = {
      email: `volunteer.${timestamp}@test.com`,
      password: 'TestPass123!',
      firstName: 'John',
      lastName: 'Volunteer',
      userType: 'volunteer',
      securityQuestions: [
        { question: "What is your mother's maiden name?", answer: 'TestAnswer1' },
        { question: 'What city were you born in?', answer: 'TestCity' },
        { question: "What was your first pet's name?", answer: 'TestPet' },
      ],
    };
    await use(user);
  },

  orgAdminUser: async ({}, use: (value: TestUser) => Promise<void>) => {
    const timestamp = Date.now();
    const user: TestUser = {
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
    await use(user);
  },
});

export const expect = baseExpect;

/**
 * Test helper functions
 */
export class TestHelpers {
  constructor(
    private page: Page,
    private apiBaseURL: string
  ) {}

  /**
   * Register a new user via API
   */
  async registerUser(
    user: TestUser
  ): Promise<{ accessToken: string; refreshToken: string; userId: string }> {
    const response = await this.page.request.post(`${this.apiBaseURL}/auth/register`, {
      data: {
        email: user.email,
        password: user.password,
        first_name: user.firstName,
        last_name: user.lastName,
        user_type: user.userType,
        security_questions: user.securityQuestions,
      },
    });

    expect(response.status()).toBe(201);
    const data = await response.json();
    return {
      accessToken: data.access_token,
      refreshToken: data.refresh_token,
      userId: data.user.id,
    };
  }

  /**
   * Login user via API
   */
  async loginUser(
    email: string,
    password: string
  ): Promise<{ accessToken: string; refreshToken: string }> {
    const response = await this.page.request.post(`${this.apiBaseURL}/auth/login`, {
      data: { email, password },
    });

    expect(response.status()).toBe(200);
    const data = await response.json();
    return {
      accessToken: data.access_token,
      refreshToken: data.refresh_token,
    };
  }

  /**
   * Set authentication tokens in browser storage
   */
  async setAuthTokens(accessToken: string, refreshToken: string) {
    await this.page.evaluate(
      ({ access, refresh }: { access: string; refresh: string }) => {
        localStorage.setItem('access_token', access);
        localStorage.setItem('refresh_token', refresh);
      },
      { access: accessToken, refresh: refreshToken }
    );
  }

  /**
   * Clear authentication tokens from browser storage
   */
  async clearAuthTokens() {
    await this.page.evaluate(() => {
      localStorage.removeItem('access_token');
      localStorage.removeItem('refresh_token');
    });
  }

  /**
   * Wait for page to finish loading
   */
  async waitForPageLoad() {
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Fill form field by label or placeholder
   */
  async fillFormField(label: string, value: string) {
    const field = this.page.getByLabel(label).or(this.page.getByPlaceholder(label));
    await field.fill(value);
  }

  /**
   * Select dropdown option by label
   */
  async selectOption(label: string, value: string) {
    await this.page.getByLabel(label).selectOption(value);
  }

  /**
   * Click button by text or test ID
   */
  async clickButton(text: string) {
    const button = this.page
      .getByRole('button', { name: text })
      .or(this.page.getByTestId(text.toLowerCase().replace(/\s+/g, '-')));
    await button.click();
  }

  /**
   * Wait for notification to appear
   */
  async waitForNotification(message?: string) {
    const notification = this.page
      .locator('[data-testid="notification"]')
      .or(this.page.locator('[role="alert"]'));
    await notification.waitFor({ state: 'visible' });
    if (message) {
      await expect(notification).toContainText(message);
    }
    return notification;
  }

  /**
   * Navigate and wait for page load
   */
  async goto(path: string) {
    await this.page.goto(path);
    await this.waitForPageLoad();
  }

  /**
   * Check if user is logged in by checking for auth tokens
   */
  async isLoggedIn(): Promise<boolean> {
    const hasToken = await this.page.evaluate(() => {
      return localStorage.getItem('access_token') !== null;
    });
    return hasToken;
  }

  /**
   * Generate unique test data
   */
  generateTestData(prefix: string): string {
    return `${prefix}_${Date.now()}_${Math.random().toString(36).substring(7)}`;
  }

  /**
   * Create organization via API (for authenticated org admin)
   */
  async createOrganization(accessToken: string, orgData: any): Promise<{ organizationId: string }> {
    const response = await this.page.request.post(`${this.apiBaseURL}/organizations`, {
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
      data: orgData,
    });

    expect(response.status()).toBe(201);
    const data = await response.json();
    return { organizationId: data.id };
  }

  /**
   * Create opportunity via API (for authenticated coordinator)
   */
  async createOpportunity(accessToken: string, oppData: any): Promise<{ opportunityId: string }> {
    const response = await this.page.request.post(`${this.apiBaseURL}/opportunities`, {
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
      data: oppData,
    });

    expect(response.status()).toBe(201);
    const data = await response.json();
    return { opportunityId: data.id };
  }

  /**
   * Measure performance timing
   */
  async measurePerformance(action: () => Promise<void>): Promise<number> {
    const startTime = Date.now();
    await action();
    const endTime = Date.now();
    return endTime - startTime;
  }

  /**
   * Take screenshot with timestamp
   */
  async takeScreenshot(name: string) {
    await this.page.screenshot({
      path: `screenshots/${name}_${Date.now()}.png`,
      fullPage: true,
    });
  }
}

/**
 * Create test helpers instance
 */
export function createHelpers(page: Page, apiBaseURL: string): TestHelpers {
  return new TestHelpers(page, apiBaseURL);
}
