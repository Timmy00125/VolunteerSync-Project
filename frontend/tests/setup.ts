/**
 * Jest Test Setup
 *
 * This file runs before each test suite.
 * Use it to configure testing libraries and global test utilities.
 */

// Only import jest-dom for tests that need it (component tests)
// For unit tests without DOM, this file can be minimal

// Reset mocks before each test
beforeEach(() => {
  jest.clearAllMocks();
});
