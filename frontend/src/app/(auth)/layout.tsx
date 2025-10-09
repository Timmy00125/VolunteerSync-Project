/**
 * Auth Layout
 *
 * Layout for authentication pages (register, login, password reset)
 * Provides minimal layout with no navigation bars
 */

import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Authentication - VolunteerSync',
  description: 'Sign in or create an account',
};

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return <>{children}</>;
}
