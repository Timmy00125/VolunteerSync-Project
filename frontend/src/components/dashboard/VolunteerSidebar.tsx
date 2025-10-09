'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { cn } from '@/lib/utils';
import { LayoutDashboard, Search, Calendar, TrendingUp, User, Heart } from 'lucide-react';
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
    href: '/volunteer',
    icon: LayoutDashboard,
    description: 'Overview of your volunteer activity',
  },
  {
    title: 'Find Opportunities',
    href: '/volunteer/opportunities',
    icon: Search,
    description: 'Discover volunteer events',
  },
  {
    title: 'My Events',
    href: '/volunteer/events',
    icon: Calendar,
    description: 'Your registered events',
  },
  {
    title: 'Impact',
    href: '/volunteer/impact',
    icon: TrendingUp,
    description: 'Track your volunteer impact',
  },
  {
    title: 'Profile',
    href: '/volunteer/profile',
    icon: User,
    description: 'Manage your volunteer profile',
  },
];

/**
 * VolunteerSidebar Component
 *
 * Navigation sidebar for the volunteer dashboard with:
 * - Active route highlighting
 * - Icon-based navigation
 * - Responsive design with scroll support
 * - Accessible navigation structure
 *
 * Used in the volunteer dashboard layout to provide consistent navigation.
 */
export function VolunteerSidebar() {
  const pathname = usePathname();

  return (
    <div className="flex h-full flex-col border-r bg-background">
      {/* Logo/Brand Section */}
      <div className="flex h-16 items-center gap-2 border-b px-6">
        <Heart className="h-6 w-6 text-primary" />
        <div className="flex flex-col">
          <span className="text-lg font-semibold">VolunteerSync</span>
          <span className="text-xs text-muted-foreground">Volunteer Portal</span>
        </div>
      </div>

      {/* Navigation Section */}
      <ScrollArea className="flex-1 px-3 py-4">
        <nav className="flex flex-col gap-1">
          {navigationItems.map((item) => {
            const isActive =
              pathname === item.href ||
              (item.href !== '/volunteer' && pathname.startsWith(item.href));
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

      {/* Footer Section (optional) */}
      <div className="border-t p-4">
        <p className="text-xs text-muted-foreground">Making an impact, one event at a time.</p>
      </div>
    </div>
  );
}
