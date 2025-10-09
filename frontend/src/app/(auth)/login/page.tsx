'use client';

/**
 * Login Page
 *
 * User authentication form with:
 * - Email and password input fields
 * - Remember me checkbox for persistent sessions
 * - Link to password reset flow
 * - Client-side validation with React Hook Form + Zod
 * - API integration to POST /auth/login
 * - Rate limiting aware (5 login attempts per 15 minutes per IP)
 *
 * Task: T104
 * Requirements: FR-001 (Authentication), NF-005 (Rate Limiting)
 */

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import Link from 'next/link';
import { loginSchema, type LoginFormData } from '@/lib/validations/auth';
import { post } from '@/lib/api/client';
import { useAuthStore } from '@/store/auth-store';
import type { AuthResponse } from '@/lib/api/types';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card } from '@/components/ui/card';

export default function LoginPage() {
  const router = useRouter();
  const { login } = useAuthStore();
  const [isLoading, setIsLoading] = useState(false);
  const [apiError, setApiError] = useState<string | null>(null);
  const [rememberMe, setRememberMe] = useState(false);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      email: '',
      password: '',
    },
  });

  // Check for password reset success message
  useEffect(() => {
    if (typeof window === 'undefined') return;

    const resetSuccess = sessionStorage.getItem('password_reset_success');
    if (resetSuccess === 'true') {
      setSuccessMessage('Password reset successful! You can now log in with your new password.');
      sessionStorage.removeItem('password_reset_success');
    }
  }, []);

  // Handle form submission
  const onSubmit = async (data: LoginFormData) => {
    setIsLoading(true);
    setApiError(null);

    try {
      // Call login API
      const response = await post<AuthResponse>('/auth/login', data, {
        skipAuth: true,
      });

      // Store user and tokens in auth store
      // The auth store persists to localStorage automatically
      login(response.user, response.tokens);

      // Optional: Set longer expiration for "remember me"
      // This is a placeholder for future implementation
      // In production, you might want to adjust token expiration on the backend
      if (rememberMe) {
        // Could send a flag to backend for extended session
        // For now, the persist middleware in auth store handles persistence
        console.log('Remember me enabled - session will persist');
      }

      // Redirect based on user type
      if (response.user.user_type === 'volunteer') {
        router.push('/volunteer'); // Volunteer dashboard
      } else if (response.user.user_type === 'organization_admin') {
        router.push('/organization'); // Organization dashboard
      } else {
        // Fallback for other user types (e.g., super_admin)
        router.push('/dashboard');
      }
    } catch (error: any) {
      console.error('Login failed:', error);

      // Handle specific error codes
      if (error.status_code === 401) {
        setApiError('Invalid email or password. Please check your credentials and try again.');
      } else if (error.status_code === 429) {
        setApiError(
          'Too many login attempts. Please wait 15 minutes before trying again. (Rate limit: 5 attempts per 15 minutes)'
        );
      } else if (error.status_code === 400) {
        setApiError(error.message || 'Invalid input. Please check your email and password.');
      } else if (error.status_code === 403) {
        setApiError(
          'Your account has been suspended or deactivated. Please contact support for assistance.'
        );
      } else {
        setApiError('Login failed. Please try again. If the problem persists, contact support.');
      }
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50 px-4 py-12 sm:px-6 lg:px-8">
      <Card className="w-full max-w-md p-8">
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
          {/* Header */}
          <div className="text-center">
            <h1 className="text-3xl font-bold tracking-tight text-gray-900">Welcome Back</h1>
            <p className="mt-2 text-sm text-gray-600">Sign in to your VolunteerSync account</p>
          </div>

          {/* Success Message Display */}
          {successMessage && (
            <div className="rounded-md bg-green-50 p-4">
              <div className="flex">
                <div className="flex-shrink-0">
                  <svg className="h-5 w-5 text-green-400" fill="currentColor" viewBox="0 0 20 20">
                    <path
                      fillRule="evenodd"
                      d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                      clipRule="evenodd"
                    />
                  </svg>
                </div>
                <div className="ml-3">
                  <p className="text-sm font-medium text-green-800">{successMessage}</p>
                </div>
              </div>
            </div>
          )}

          {/* API Error Display */}
          {apiError && (
            <div className="rounded-md bg-red-50 p-4">
              <div className="flex">
                <div className="flex-shrink-0">
                  <svg className="h-5 w-5 text-red-400" fill="currentColor" viewBox="0 0 20 20">
                    <path
                      fillRule="evenodd"
                      d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                      clipRule="evenodd"
                    />
                  </svg>
                </div>
                <div className="ml-3">
                  <p className="text-sm font-medium text-red-800">{apiError}</p>
                </div>
              </div>
            </div>
          )}

          {/* Form Fields */}
          <div className="space-y-4">
            {/* Email */}
            <div>
              <Label htmlFor="email">Email Address</Label>
              <Input
                id="email"
                type="email"
                autoComplete="email"
                placeholder="you@example.com"
                {...register('email')}
                className={errors.email ? 'border-red-500' : ''}
                disabled={isLoading}
              />
              {errors.email && <p className="mt-1 text-sm text-red-600">{errors.email.message}</p>}
            </div>

            {/* Password */}
            <div>
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                type="password"
                autoComplete="current-password"
                placeholder="••••••••"
                {...register('password')}
                className={errors.password ? 'border-red-500' : ''}
                disabled={isLoading}
              />
              {errors.password && (
                <p className="mt-1 text-sm text-red-600">{errors.password.message}</p>
              )}
            </div>
          </div>

          {/* Remember Me & Forgot Password */}
          <div className="flex items-center justify-between">
            <div className="flex items-center">
              <input
                id="remember-me"
                name="remember-me"
                type="checkbox"
                checked={rememberMe}
                onChange={(e) => setRememberMe(e.target.checked)}
                disabled={isLoading}
                className="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
              />
              <label htmlFor="remember-me" className="ml-2 block text-sm text-gray-900">
                Remember me
              </label>
            </div>

            <div className="text-sm">
              <Link
                href="/reset-password"
                className="font-medium text-blue-600 hover:text-blue-500"
              >
                Forgot password?
              </Link>
            </div>
          </div>

          {/* Submit Button */}
          <div>
            <Button type="submit" className="w-full" disabled={isLoading}>
              {isLoading ? (
                <span className="flex items-center justify-center">
                  <svg
                    className="-ml-1 mr-3 h-5 w-5 animate-spin text-white"
                    xmlns="http://www.w3.org/2000/svg"
                    fill="none"
                    viewBox="0 0 24 24"
                  >
                    <circle
                      className="opacity-25"
                      cx="12"
                      cy="12"
                      r="10"
                      stroke="currentColor"
                      strokeWidth="4"
                    ></circle>
                    <path
                      className="opacity-75"
                      fill="currentColor"
                      d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                    ></path>
                  </svg>
                  Signing in...
                </span>
              ) : (
                'Sign in'
              )}
            </Button>
          </div>

          {/* Divider */}
          <div className="relative">
            <div className="absolute inset-0 flex items-center">
              <div className="w-full border-t border-gray-300" />
            </div>
            <div className="relative flex justify-center text-sm">
              <span className="bg-white px-2 text-gray-500">Don't have an account?</span>
            </div>
          </div>

          {/* Register Link */}
          <div className="text-center">
            <Link href="/register" className="font-medium text-blue-600 hover:text-blue-500">
              Create a new account
            </Link>
          </div>
        </form>
      </Card>
    </div>
  );
}
