'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { cn } from '@/lib/utils';
import { LayoutDashboard, Calendar, Users, BarChart3, Settings, Building2 } from 'lucide-react';
import { ScrollArea } from '@/components/ui/scroll-area';

interface NavItem {
  title: string;
  href: string;
  icon: React.ComponentType<{ className?: string }>;
  description: string;
}

const navigationItems: NavItem[] = [
  {
    title: 'Dashboard',
    href: '/organization',
    icon: LayoutDashboard,
    description: 'Overview of organization activity',
  },
  {
    title: 'Opportunities',
    href: '/organization/opportunities',
    icon: Calendar,
    description: 'Manage volunteer opportunities',
  },
  {
    title: 'Team',
    href: '/organization/team',
    icon: Users,
    description: 'Manage team members and roles',
  },
  {
    title: 'Analytics',
    href: '/organization/analytics',
    icon: BarChart3,
    description: 'View organization analytics',
  },
  {
    title: 'Settings',
    href: '/organization/settings',
    icon: Settings,
    description: 'Organization settings',
  },
];

/**
 * OrganizationSidebar Component
 *
 * Navigation sidebar for the organization dashboard with:
 * - Active route highlighting
 * - Icon-based navigation
 * - Organization-specific sections (Opportunities, Team, Analytics, Settings)
 * - Responsive design with scroll support
 * - Accessible navigation structure
 *
 * Used in the organization dashboard layout to provide consistent navigation
 * for organization administrators and coordinators.
 */
export function OrganizationSidebar() {
  const pathname = usePathname();

  return (
    <div className="flex h-full flex-col border-r bg-background">
      {/* Logo/Brand Section */}
      <div className="flex h-16 items-center gap-2 border-b px-6">
        <Building2 className="h-6 w-6 text-primary" />
        <div className="flex flex-col">
          <span className="text-lg font-semibold">VolunteerSync</span>
          <span className="text-xs text-muted-foreground">Organization Portal</span>
        </div>
      </div>

      {/* Navigation Section */}
      <ScrollArea className="flex-1 px-3 py-4">
        <nav className="flex flex-col gap-1">
          {navigationItems.map((item) => {
            const isActive =
              pathname === item.href ||
              (item.href !== '/organization' && pathname.startsWith(item.href));
            const Icon = item.icon;

            return (
              <Link
                key={item.href}
                href={item.href}
                className={cn(
                  'flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors',
                  'hover:bg-accent hover:text-accent-foreground',
                  'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
                  isActive ? 'bg-accent text-accent-foreground' : 'text-muted-foreground'
                )}
                aria-current={isActive ? 'page' : undefined}
              >
                <Icon className="h-5 w-5 shrink-0" aria-hidden="true" />
                <span>{item.title}</span>
              </Link>
            );
          })}
        </nav>
      </ScrollArea>

      {/* Footer Section */}
      <div className="border-t p-4">
        <p className="text-xs text-muted-foreground">
          Empowering volunteers, strengthening communities.
        </p>
      </div>
    </div>
  );
}
