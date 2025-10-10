'use client';

import { Bell, LogOut, Settings, User } from 'lucide-react';
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
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/store/auth-store';
import { post } from '@/lib/api/client';

interface VolunteerHeaderProps {
  /**
   * Number of unread notifications to display in badge
   */
  unreadCount?: number;
}

/**
 * VolunteerHeader Component
 *
 * Top header bar for the volunteer dashboard featuring:
 * - Notifications bell with unread count badge
 * - User menu with avatar and dropdown
 * - Quick access to profile settings
 * - Logout functionality
 *
 * This component is used in the volunteer dashboard layout to provide
 * consistent navigation and user controls across all volunteer pages.
 */
export function VolunteerHeader({ unreadCount = 0 }: VolunteerHeaderProps) {
  const router = useRouter();
  const { user, logout } = useAuthStore();

  // Get user data from auth store
  const displayUser = user
    ? {
        name: `${user.first_name} ${user.last_name}`.trim() || user.email,
        email: user.email,
        avatar: user.profile_photo_url,
      }
    : {
        name: 'User',
        email: 'user@example.com',
      };

  // Generate initials from name for avatar fallback
  const initials = displayUser.name
    .split(' ')
    .map((n: string) => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2);

  const handleLogout = async () => {
    try {
      // Call logout API endpoint
      await post('/auth/logout', {});
    } catch (error) {
      console.error('Logout API call failed:', error);
      // Continue with logout even if API call fails
    } finally {
      // Clear auth state from store (and localStorage)
      logout();

      // Clear the auth cookie used by middleware
      document.cookie = 'auth-user-type=; path=/; expires=Thu, 01 Jan 1970 00:00:01 GMT';

      // Redirect to login page
      router.push('/login');
    }
  };

  return (
    <header className="sticky top-0 z-50 flex h-16 items-center justify-between border-b bg-background px-6">
      {/* Left Section - Page Title (can be customized per page) */}
      <div className="flex items-center gap-4">
        <h1 className="text-xl font-semibold">Dashboard</h1>
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
                  <Link href="/volunteer/notifications" className="flex flex-col items-start py-3">
                    <span className="font-medium">New volunteer opportunity</span>
                    <span className="text-xs text-muted-foreground">
                      Environmental cleanup event this weekend
                    </span>
                  </Link>
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem asChild>
                  <Link href="/volunteer/notifications" className="justify-center text-center">
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
                <AvatarImage src={displayUser.avatar} alt={displayUser.name} />
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
              <Link href="/volunteer/settings" className="flex items-center gap-2">
                <Settings className="h-4 w-4" />
                <span>Settings</span>
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
