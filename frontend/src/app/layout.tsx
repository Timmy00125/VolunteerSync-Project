import type { Metadata } from 'next';
import { Geist, Geist_Mono } from 'next/font/google';
import './globals.css';
import { Providers } from '@/components/providers';

const geistSans = Geist({
  variable: '--font-geist-sans',
  subsets: ['latin'],
});

const geistMono = Geist_Mono({
  variable: '--font-geist-mono',
  subsets: ['latin'],
});

export const metadata: Metadata = {
  title: 'VolunteerSync - Connect. Serve. Impact.',
  description:
    'A comprehensive platform connecting nonprofit organizations with volunteers, enabling seamless opportunity discovery, event management, and impact tracking.',
  keywords:
    'volunteer, nonprofit, volunteering, community service, social impact, volunteer management',
  authors: [{ name: 'VolunteerSync Team' }],
  openGraph: {
    title: 'VolunteerSync - Connect. Serve. Impact.',
    description: 'Connect with meaningful volunteer opportunities and track your community impact.',
    type: 'website',
  },
};

/**
 * Root Layout Component
 *
 * Provides the base HTML structure and global configuration for the entire application.
 *
 * Features:
 * - HTML structure with proper lang attribute
 * - Font loading with Geist Sans and Geist Mono
 * - Global styles from globals.css
 * - React Query Provider for server state management
 * - Auth context for authentication state
 * - Proper metadata for SEO
 *
 * This layout wraps all pages in the application.
 */
export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
        suppressHydrationWarning
      >
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
