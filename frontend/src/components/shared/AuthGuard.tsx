/**
 * Auth Guard Component
 *
 * Client-side route protection that wraps protected pages and ensures
 * users are authenticated before rendering content.
 *
 * Features:
 * - Checks authentication state from Zustand store
 * - Redirects to login if not authenticated
 * - Shows loading state while checking auth
 * - Validates user type for role-based access
 *
 * Usage:
 * Wrap any page content that requires authentication:
 * ```tsx
 * <AuthGuard requiredUserType="volunteer">
 *   <YourProtectedContent />
 * </AuthGuard>
 * ```
 */

'use client';

import { useEffect, useState } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import { useAuthStore } from '@/store/auth-store';
import { Loader2 } from 'lucide-react';

interface AuthGuardProps {
  children: React.ReactNode;
  /**
   * Required user type(s) to access this route
   * If not specified, any authenticated user can access
   */
  requiredUserType?: 'volunteer' | 'organization_admin' | 'super_admin' | string[];
  /**
   * Custom loading component
   */
  loadingComponent?: React.ReactNode;
  /**
   * Custom redirect path (defaults to /login)
   */
  redirectTo?: string;
}

/**
 * Loading spinner component
 */
function DefaultLoadingComponent() {
  return (
    <div className="flex min-h-screen items-center justify-center">
      <div className="text-center">
        <Loader2 className="mx-auto h-12 w-12 animate-spin text-primary" />
        <p className="mt-4 text-sm text-muted-foreground">Checking authentication...</p>
      </div>
    </div>
  );
}

/**
 * AuthGuard Component
 *
 * Protects routes by checking authentication state and user permissions.
 * Automatically redirects to login if user is not authenticated or doesn't
 * have the required user type.
 */
export function AuthGuard({
  children,
  requiredUserType,
  loadingComponent,
  redirectTo = '/login',
}: AuthGuardProps) {
  const router = useRouter();
  const pathname = usePathname();
  const { user, isAuthenticated, isLoading } = useAuthStore();
  const [isChecking, setIsChecking] = useState(true);

  useEffect(() => {
    // Wait for auth store to finish loading from localStorage
    if (isLoading) {
      return;
    }

    // Check if user is authenticated
    if (!isAuthenticated || !user) {
      console.log('AuthGuard: User not authenticated, redirecting to login');
      router.push(`${redirectTo}?redirect=${encodeURIComponent(pathname)}`);
      return;
    }

    // Check user type if required
    if (requiredUserType) {
      const allowedTypes = Array.isArray(requiredUserType) ? requiredUserType : [requiredUserType];

      // Super admins have access to everything
      const hasAccess = user.user_type === 'super_admin' || allowedTypes.includes(user.user_type);

      if (!hasAccess) {
        console.log(
          `AuthGuard: User type ${user.user_type} not allowed. Required: ${allowedTypes.join(', ')}`
        );
        // Redirect to appropriate dashboard based on user type
        if (user.user_type === 'volunteer') {
          router.push('/volunteer');
        } else if (user.user_type === 'organization_admin') {
          router.push('/organization');
        } else {
          router.push('/');
        }
        return;
      }
    }

    // All checks passed
    setIsChecking(false);
  }, [isAuthenticated, isLoading, user, requiredUserType, router, pathname, redirectTo]);

  // Show loading state while checking authentication
  if (isLoading || isChecking) {
    return loadingComponent || <DefaultLoadingComponent />;
  }

  // Don't render children if not authenticated (redirect will happen)
  if (!isAuthenticated || !user) {
    return null;
  }

  // Check user type before rendering
  if (requiredUserType) {
    const allowedTypes = Array.isArray(requiredUserType) ? requiredUserType : [requiredUserType];
    const hasAccess = user.user_type === 'super_admin' || allowedTypes.includes(user.user_type);

    if (!hasAccess) {
      return null;
    }
  }

  // Render protected content
  return <>{children}</>;
}
