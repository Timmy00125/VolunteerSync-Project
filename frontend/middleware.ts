/**
 * Next.js Middleware for Authentication
 *
 * Protects routes based on authentication status and user type.
 * Runs on the Edge Runtime before rendering pages.
 *
 * Protected Routes:
 * - /volunteer/* - Requires volunteer or admin user
 * - /organization/* - Requires organization_admin or super_admin user
 *
 * Auth Routes (redirect if logged in):
 * - /login
 * - /register
 * - /reset-password/*
 */

import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

// Define route patterns
const AUTH_ROUTES = ['/login', '/register', '/reset-password'];
const VOLUNTEER_ROUTES = ['/volunteer'];
const ORGANIZATION_ROUTES = ['/organization'];

/**
 * Check if the path matches any of the given route patterns
 */
function matchesRoute(pathname: string, routes: string[]): boolean {
  return routes.some((route) => pathname.startsWith(route));
}

/**
 * Get auth state from localStorage (via cookie fallback for middleware)
 * Note: Middleware can't access localStorage, so we'll use a different approach
 */
function getAuthFromCookies(request: NextRequest): {
  isAuthenticated: boolean;
  userType: string | null;
} {
  // Try to get auth state from a cookie (we'll need to set this on login)
  const authCookie = request.cookies.get('auth-user-type');

  if (authCookie) {
    return {
      isAuthenticated: true,
      userType: authCookie.value,
    };
  }

  return {
    isAuthenticated: false,
    userType: null,
  };
}

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;
  const { isAuthenticated, userType } = getAuthFromCookies(request);

  // Allow access to public routes (home page, public opportunity listings, etc.)
  if (pathname === '/' || pathname.startsWith('/opportunities/')) {
    return NextResponse.next();
  }

  // Handle auth routes (login, register, reset-password)
  if (matchesRoute(pathname, AUTH_ROUTES)) {
    // If user is already authenticated, redirect to appropriate dashboard
    if (isAuthenticated) {
      const redirectUrl = new URL(
        userType === 'volunteer' ? '/volunteer' : '/organization',
        request.url
      );
      return NextResponse.redirect(redirectUrl);
    }
    // Allow access to auth pages if not authenticated
    return NextResponse.next();
  }

  // Handle protected volunteer routes
  if (matchesRoute(pathname, VOLUNTEER_ROUTES)) {
    if (!isAuthenticated) {
      // Not authenticated - redirect to login
      const loginUrl = new URL('/login', request.url);
      loginUrl.searchParams.set('redirect', pathname);
      return NextResponse.redirect(loginUrl);
    }

    // Check if user has volunteer access
    if (userType !== 'volunteer' && userType !== 'super_admin') {
      // Wrong user type - redirect to their dashboard
      const redirectUrl = new URL('/organization', request.url);
      return NextResponse.redirect(redirectUrl);
    }

    return NextResponse.next();
  }

  // Handle protected organization routes
  if (matchesRoute(pathname, ORGANIZATION_ROUTES)) {
    if (!isAuthenticated) {
      // Not authenticated - redirect to login
      const loginUrl = new URL('/login', request.url);
      loginUrl.searchParams.set('redirect', pathname);
      return NextResponse.redirect(loginUrl);
    }

    // Check if user has organization admin access
    if (userType !== 'organization_admin' && userType !== 'super_admin') {
      // Wrong user type - redirect to their dashboard
      const redirectUrl = new URL('/volunteer', request.url);
      return NextResponse.redirect(redirectUrl);
    }

    return NextResponse.next();
  }

  // Allow all other routes
  return NextResponse.next();
}

// Configure which routes the middleware should run on
export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - api (API routes)
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     * - public files (images, etc.)
     */
    '/((?!api|_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp)$).*)',
  ],
};
