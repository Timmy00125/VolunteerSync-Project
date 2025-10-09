'use client';

import { Bell, LogOut, Settings, User, Building2, ChevronDown } from 'lucide-react';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import Link from 'next/link';

interface Organization {
  id: string;
  name: string;
  logo?: string;
}

interface OrganizationHeaderProps {
  /**
   * Number of unread notifications to display in badge
   */
  unreadCount?: number;
  /**
   * User information to display in the header
   */
  user?: {
    name: string;
    email: string;
    avatar?: string;
  };
  /**
   * Current organization being managed
   */
  currentOrganization?: Organization;
  /**
   * List of organizations the user is a member of (for org switcher)
   */
  organizations?: Organization[];
}

/**
 * OrganizationHeader Component
 *
 * Top header bar for the organization dashboard featuring:
 * - Organization switcher (if member of multiple organizations)
 * - Notifications bell with unread count badge
 * - User menu with avatar and dropdown
 * - Quick access to profile settings
 * - Logout functionality
 *
 * This component is used in the organization dashboard layout to provide
 * consistent navigation and controls across all organization pages.
 *
 * @example
 * ```tsx
 * <OrganizationHeader
 *   unreadCount={3}
 *   user={{ name: "Jane Doe", email: "jane@example.com" }}
 *   currentOrganization={{ id: "1", name: "Community Helpers" }}
 *   organizations={[
 *     { id: "1", name: "Community Helpers" },
 *     { id: "2", name: "Green Earth Initiative" }
 *   ]}
 * />
 * ```
 */
export function OrganizationHeader({
  unreadCount = 0,
  user,
  currentOrganization,
  organizations = [],
}: OrganizationHeaderProps) {
  // Default user fallback for development
  const displayUser = user || {
    name: 'Organization Admin',
    email: 'admin@example.com',
  };

  // Default organization fallback
  const displayOrg = currentOrganization || {
    id: '1',
    name: 'My Organization',
  };

  // Generate initials from name for avatar fallback
  const initials = displayUser.name
    .split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2);

  // Generate initials for organization logo fallback
  const orgInitials = displayOrg.name
    .split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2);

  const handleLogout = () => {
    // TODO: Implement logout logic with auth context
    console.log('Logout clicked');
  };

  const handleOrganizationSwitch = (orgId: string) => {
    // TODO: Implement organization switch logic with context/state management
    console.log('Switching to organization:', orgId);
  };

  // Show organization switcher only if user is member of multiple orgs
  const showOrgSwitcher = organizations.length > 1;

  return (
    <header className="sticky top-0 z-50 flex h-16 items-center justify-between border-b bg-background px-6">
      {/* Left Section - Organization Switcher */}
      <div className="flex items-center gap-4">
        {showOrgSwitcher ? (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                variant="ghost"
                className="flex items-center gap-2 px-2"
                aria-label="Switch organization"
              >
                <Avatar className="h-8 w-8">
                  <AvatarImage src={displayOrg.logo} alt={displayOrg.name} />
                  <AvatarFallback className="bg-primary text-primary-foreground">
                    {orgInitials}
                  </AvatarFallback>
                </Avatar>
                <div className="flex flex-col items-start text-left">
                  <span className="text-sm font-medium">{displayOrg.name}</span>
                  <span className="text-xs text-muted-foreground">Switch organization</span>
                </div>
                <ChevronDown className="h-4 w-4 text-muted-foreground" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="start" className="w-64">
              <DropdownMenuLabel>Your Organizations</DropdownMenuLabel>
              <DropdownMenuSeparator />
              {organizations.map((org) => (
                <DropdownMenuItem
                  key={org.id}
                  onClick={() => handleOrganizationSwitch(org.id)}
                  className="flex items-center gap-2 py-2"
                >
                  <Avatar className="h-6 w-6">
                    <AvatarImage src={org.logo} alt={org.name} />
                    <AvatarFallback className="bg-primary/10 text-xs">
                      {org.name
                        .split(' ')
                        .map((n) => n[0])
                        .join('')
                        .toUpperCase()
                        .slice(0, 2)}
                    </AvatarFallback>
                  </Avatar>
                  <span className="flex-1">{org.name}</span>
                  {org.id === displayOrg.id && (
                    <Badge variant="secondary" className="ml-auto text-xs">
                      Current
                    </Badge>
                  )}
                </DropdownMenuItem>
              ))}
              <DropdownMenuSeparator />
              <DropdownMenuItem asChild>
                <Link href="/organization/new" className="flex items-center gap-2">
                  <Building2 className="h-4 w-4" />
                  <span>Create New Organization</span>
                </Link>
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        ) : (
          // Single organization - just show name with icon
          <div className="flex items-center gap-2">
            <Avatar className="h-8 w-8">
              <AvatarImage src={displayOrg.logo} alt={displayOrg.name} />
              <AvatarFallback className="bg-primary text-primary-foreground">
                {orgInitials}
              </AvatarFallback>
            </Avatar>
            <div className="flex flex-col">
              <span className="text-sm font-medium">{displayOrg.name}</span>
            </div>
          </div>
        )}
      </div>

      {/* Right Section - Notifications & User Menu */}
      <div className="flex items-center gap-4">
        {/* Notifications Bell */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button
              variant="ghost"
              size="icon"
              className="relative"
              aria-label={`Notifications${unreadCount > 0 ? `, ${unreadCount} unread` : ''}`}
            >
              <Bell className="h-5 w-5" />
              {unreadCount > 0 && (
                <Badge
                  variant="destructive"
                  className="absolute -right-1 -top-1 flex h-5 min-w-[20px] items-center justify-center rounded-full px-1 text-xs"
                >
                  {unreadCount > 99 ? '99+' : unreadCount}
                </Badge>
              )}
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-80">
            <DropdownMenuLabel>Notifications</DropdownMenuLabel>
            <DropdownMenuSeparator />
            {unreadCount > 0 ? (
              <>
                {/* TODO: Replace with actual notifications from API */}
                <DropdownMenuItem asChild>
                  <Link
                    href="/organization/notifications"
                    className="flex flex-col items-start py-3"
                  >
                    <span className="font-medium">New volunteer registration</span>
                    <span className="text-xs text-muted-foreground">
                      5 volunteers signed up for tomorrow's event
                    </span>
                  </Link>
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem asChild>
                  <Link href="/organization/notifications" className="justify-center text-center">
                    View all notifications
                  </Link>
                </DropdownMenuItem>
              </>
            ) : (
              <div className="py-6 text-center text-sm text-muted-foreground">
                No new notifications
              </div>
            )}
          </DropdownMenuContent>
        </DropdownMenu>

        {/* User Menu */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" className="flex items-center gap-2 px-2" aria-label="User menu">
              <Avatar className="h-8 w-8">
                <AvatarImage src={user?.avatar} alt={displayUser.name} />
                <AvatarFallback>{initials}</AvatarFallback>
              </Avatar>
              <div className="hidden flex-col items-start text-left md:flex">
                <span className="text-sm font-medium">{displayUser.name}</span>
                <span className="text-xs text-muted-foreground">{displayUser.email}</span>
              </div>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-56">
            <DropdownMenuLabel>My Account</DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem asChild>
              <Link href="/volunteer/profile" className="flex items-center gap-2">
                <User className="h-4 w-4" />
                <span>Profile</span>
              </Link>
            </DropdownMenuItem>
            <DropdownMenuItem asChild>
              <Link href="/organization/settings" className="flex items-center gap-2">
                <Settings className="h-4 w-4" />
                <span>Organization Settings</span>
              </Link>
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={handleLogout}
              className="flex items-center gap-2 text-destructive focus:text-destructive"
            >
              <LogOut className="h-4 w-4" />
              <span>Log out</span>
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  );
}
