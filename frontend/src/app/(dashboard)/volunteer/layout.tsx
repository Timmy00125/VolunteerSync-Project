import { VolunteerSidebar } from '@/components/dashboard/VolunteerSidebar';
import { VolunteerHeader } from '@/components/dashboard/VolunteerHeader';

/**
 * Volunteer Dashboard Layout
 *
 * Provides the layout structure for all volunteer dashboard pages.
 *
 * Features:
 * - Left sidebar navigation with active state highlighting
 * - Top header with notifications bell and user menu
 * - Main content area with proper spacing and scroll behavior
 * - Responsive design (sidebar collapses on mobile)
 * - Consistent styling across all volunteer pages
 *
 * Navigation Structure:
 * - Dashboard: Overview with impact metrics and recent activity
 * - Find Opportunities: Search and filter volunteer events
 * - My Events: Registered and past events
 * - Impact: Personal metrics, hours, and achievements
 * - Profile: Volunteer profile management
 *
 * The layout uses a flexbox structure with:
 * - Fixed sidebar (width: 16rem / 256px)
 * - Fixed header (height: 4rem / 64px)
 * - Flexible content area that fills remaining space
 *
 * @example
 * All pages under /volunteer/* will automatically use this layout:
 * - /volunteer → Dashboard page
 * - /volunteer/opportunities → Find Opportunities page
 * - /volunteer/events → My Events page
 * - etc.
 */
export default function VolunteerDashboardLayout({ children }: { children: React.ReactNode }) {
  // TODO: Fetch user data and unread notification count from auth context/API
  // For now using placeholder data
  const mockUser = {
    name: 'Jane Volunteer',
    email: 'jane@example.com',
    avatar: undefined, // Will show initials instead
  };

  const mockUnreadCount = 0; // TODO: Get from notifications API

  return (
    <div className="flex h-screen overflow-hidden bg-background">
      {/* Sidebar - Fixed width, full height */}
      <aside className="hidden w-64 md:block">
        <VolunteerSidebar />
      </aside>

      {/* Main Content Area */}
      <div className="flex flex-1 flex-col overflow-hidden">
        {/* Header - Fixed at top */}
        <VolunteerHeader user={mockUser} unreadCount={mockUnreadCount} />

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
