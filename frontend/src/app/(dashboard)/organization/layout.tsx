import { OrganizationSidebar } from '@/components/dashboard/OrganizationSidebar';
import { OrganizationHeader } from '@/components/dashboard/OrganizationHeader';

/**
 * Organization Dashboard Layout
 *
 * Provides the layout structure for all organization dashboard pages.
 *
 * Features:
 * - Left sidebar navigation with active state highlighting
 * - Top header with organization switcher (for multi-org users)
 * - Notifications bell and user menu
 * - Main content area with proper spacing and scroll behavior
 * - Responsive design (sidebar collapses on mobile)
 * - Consistent styling across all organization pages
 *
 * Navigation Structure:
 * - Dashboard: Overview with org metrics and recent activity
 * - Opportunities: Create and manage volunteer opportunities
 * - Team: Manage team members, roles, and permissions
 * - Analytics: View organization analytics and volunteer insights
 * - Settings: Organization profile, verification, and preferences
 *
 * The layout uses a flexbox structure with:
 * - Fixed sidebar (width: 16rem / 256px)
 * - Fixed header (height: 4rem / 64px)
 * - Flexible content area that fills remaining space
 *
 * @example
 * All pages under /organization/* will automatically use this layout:
 * - /organization → Dashboard page
 * - /organization/opportunities → Opportunities management
 * - /organization/team → Team management
 * - /organization/analytics → Analytics page
 * - /organization/settings → Settings page
 */
export default function OrganizationDashboardLayout({ children }: { children: React.ReactNode }) {
  // TODO: Fetch user data, current organization, and unread notification count
  // This should come from auth context/API
  const mockUser = {
    name: 'Jane Coordinator',
    email: 'jane.coordinator@example.com',
    avatar: undefined, // Will show initials instead
  };

  const mockCurrentOrg = {
    id: '1',
    name: 'Community Helpers',
    logo: undefined,
  };

  // TODO: In production, fetch this from organizations API
  // This mock shows the organization switcher feature
  const mockOrganizations = [
    {
      id: '1',
      name: 'Community Helpers',
      logo: undefined,
    },
    {
      id: '2',
      name: 'Green Earth Initiative',
      logo: undefined,
    },
  ];

  const mockUnreadCount = 0; // TODO: Get from notifications API

  return (
    <div className="flex h-screen overflow-hidden bg-background">
      {/* Sidebar - Fixed width, full height */}
      <aside className="hidden w-64 md:block">
        <OrganizationSidebar />
      </aside>

      {/* Main Content Area */}
      <div className="flex flex-1 flex-col overflow-hidden">
        {/* Header - Fixed at top */}
        <OrganizationHeader
          user={mockUser}
          currentOrganization={mockCurrentOrg}
          organizations={mockOrganizations}
          unreadCount={mockUnreadCount}
        />

        {/* Page Content - Scrollable */}
        <main className="flex-1 overflow-y-auto">
          <div className="container mx-auto p-6">{children}</div>
        </main>
      </div>

      {/* Mobile Sidebar Overlay - TODO: Implement mobile menu toggle */}
      {/* This would be a slide-in overlay for mobile devices */}
      {/* Implementation can be added in a future iteration with a hamburger menu */}
    </div>
  );
}
