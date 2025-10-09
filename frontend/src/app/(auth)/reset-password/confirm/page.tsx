'use client';

/**
 * Password Reset Confirm Page
 *
 * Final step: Set new password after security question verification.
 *
 * Flow:
 * 1. User enters new password and confirmation
 * 2. API updates password using verified_token
 * 3. On success, redirect to login with success message
 *
 * Task: T105 (Part 3 of 3)
 * Requirements: FR-003 (Password Reset), NF-006 (Password Strength)
 * API: POST /api/v1/auth/password-reset/confirm
 */

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import Link from 'next/link';
import { passwordResetConfirmSchema, type PasswordResetConfirmData } from '@/lib/validations/auth';
import { post } from '@/lib/api/client';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card } from '@/components/ui/card';

export default function ResetPasswordConfirmPage() {
  const router = useRouter();
  const [isLoading, setIsLoading] = useState(false);
  const [apiError, setApiError] = useState<string | null>(null);
  const [verifiedToken, setVerifiedToken] = useState<string>('');
  const [email, setEmail] = useState<string>('');
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
    setValue,
    watch,
  } = useForm<PasswordResetConfirmData>({
    resolver: zodResolver(passwordResetConfirmSchema),
    defaultValues: {
      verified_token: '',
      new_password: '',
      confirm_password: '',
    },
  });

  const newPassword = watch('new_password');

  // Load verified token from sessionStorage on mount
  useEffect(() => {
    if (typeof window === 'undefined') return;

    const token = sessionStorage.getItem('verified_token');
    const storedEmail = sessionStorage.getItem('reset_email');

    if (!token) {
      // No verified token found - redirect to request page
      router.push('/reset-password');
      return;
    }

    setVerifiedToken(token);
    setEmail(storedEmail || '');
    setValue('verified_token', token);
  }, [router, setValue]);

  // Handle form submission
  const onSubmit = async (data: PasswordResetConfirmData) => {
    setIsLoading(true);
    setApiError(null);

    try {
      // Call password reset confirm API
      await post(
        '/auth/password-reset/confirm',
        {
          verified_token: data.verified_token,
          new_password: data.new_password,
        },
        { skipAuth: true }
      );

      // Clean up session storage
      if (typeof window !== 'undefined') {
        sessionStorage.removeItem('verified_token');
        sessionStorage.removeItem('reset_email');
        // Set success flag for login page
        sessionStorage.setItem('password_reset_success', 'true');
      }

      // Redirect to login page with success message
      router.push('/login');
    } catch (error: any) {
      console.error('Password reset confirm failed:', error);

      // Handle specific error codes
      if (error.status_code === 400) {
        if (error.message?.includes('password')) {
          setApiError(
            'Password does not meet requirements. Must be at least 8 characters with letters and numbers.'
          );
        } else {
          setApiError(error.message || 'Invalid input. Please check your password and try again.');
        }
      } else if (error.status_code === 401 || error.status_code === 404) {
        setApiError('Password reset session expired. Please start over from the beginning.');
        // Clear session storage and redirect after delay
        setTimeout(() => {
          if (typeof window !== 'undefined') {
            sessionStorage.removeItem('verified_token');
            sessionStorage.removeItem('reset_email');
          }
          router.push('/reset-password');
        }, 3000);
      } else if (error.status_code === 429) {
        setApiError('Too many password reset attempts. Please wait before trying again.');
      } else {
        setApiError('Unable to reset password. Please try again later.');
      }
    } finally {
      setIsLoading(false);
    }
  };

  // Password strength indicator
  const getPasswordStrength = (
    password: string
  ): {
    strength: 'weak' | 'fair' | 'good' | 'strong';
    label: string;
    color: string;
  } => {
    if (!password) return { strength: 'weak', label: '', color: '' };

    let score = 0;
    if (password.length >= 8) score++;
    if (password.length >= 12) score++;
    if (/[a-z]/.test(password) && /[A-Z]/.test(password)) score++;
    if (/\d/.test(password)) score++;
    if (/[^A-Za-z0-9]/.test(password)) score++;

    if (score <= 1) return { strength: 'weak', label: 'Weak', color: 'bg-red-500' };
    if (score === 2) return { strength: 'fair', label: 'Fair', color: 'bg-yellow-500' };
    if (score === 3 || score === 4)
      return { strength: 'good', label: 'Good', color: 'bg-blue-500' };
    return { strength: 'strong', label: 'Strong', color: 'bg-green-500' };
  };

  const passwordStrength = getPasswordStrength(newPassword);

  // Show loading while checking for verified token
  if (!verifiedToken) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-gray-50">
        <div className="text-center">
          <svg
            className="mx-auto h-12 w-12 animate-spin text-blue-600"
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
          <p className="mt-4 text-sm text-gray-600">Loading...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50 px-4 py-12 sm:px-6 lg:px-8">
      <Card className="w-full max-w-md p-8">
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
          {/* Header */}
          <div className="text-center">
            <h1 className="text-3xl font-bold tracking-tight text-gray-900">Set New Password</h1>
            {email && (
              <p className="mt-2 text-sm text-gray-600">
                For account: <span className="font-medium">{email}</span>
              </p>
            )}
            <p className="mt-2 text-sm text-gray-600">
              Choose a strong password to protect your account
            </p>
          </div>

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
            {/* New Password */}
            <div>
              <Label htmlFor="new_password">New Password</Label>
              <div className="relative">
                <Input
                  id="new_password"
                  type={showPassword ? 'text' : 'password'}
                  autoComplete="new-password"
                  placeholder="••••••••"
                  {...register('new_password')}
                  className={errors.new_password ? 'border-red-500' : ''}
                  disabled={isLoading}
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-gray-700"
                >
                  {showPassword ? (
                    <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21"
                      />
                    </svg>
                  ) : (
                    <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                      />
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"
                      />
                    </svg>
                  )}
                </button>
              </div>
              {errors.new_password && (
                <p className="mt-1 text-sm text-red-600">{errors.new_password.message}</p>
              )}
              {/* Password Strength Indicator */}
              {newPassword && passwordStrength.label && (
                <div className="mt-2">
                  <div className="flex items-center justify-between text-xs text-gray-600">
                    <span>Password strength:</span>
                    <span className="font-medium">{passwordStrength.label}</span>
                  </div>
                  <div className="mt-1 h-1.5 w-full rounded-full bg-gray-200">
                    <div
                      className={`h-full rounded-full transition-all ${passwordStrength.color}`}
                      style={{
                        width:
                          passwordStrength.strength === 'weak'
                            ? '25%'
                            : passwordStrength.strength === 'fair'
                              ? '50%'
                              : passwordStrength.strength === 'good'
                                ? '75%'
                                : '100%',
                      }}
                    />
                  </div>
                </div>
              )}
            </div>

            {/* Confirm Password */}
            <div>
              <Label htmlFor="confirm_password">Confirm New Password</Label>
              <div className="relative">
                <Input
                  id="confirm_password"
                  type={showConfirmPassword ? 'text' : 'password'}
                  autoComplete="new-password"
                  placeholder="••••••••"
                  {...register('confirm_password')}
                  className={errors.confirm_password ? 'border-red-500' : ''}
                  disabled={isLoading}
                />
                <button
                  type="button"
                  onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-gray-700"
                >
                  {showConfirmPassword ? (
                    <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21"
                      />
                    </svg>
                  ) : (
                    <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                      />
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"
                      />
                    </svg>
                  )}
                </button>
              </div>
              {errors.confirm_password && (
                <p className="mt-1 text-sm text-red-600">{errors.confirm_password.message}</p>
              )}
            </div>
          </div>

          {/* Password Requirements */}
          <div className="rounded-md bg-blue-50 p-4">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-blue-400" fill="currentColor" viewBox="0 0 20 20">
                  <path
                    fillRule="evenodd"
                    d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
                    clipRule="evenodd"
                  />
                </svg>
              </div>
              <div className="ml-3">
                <p className="text-sm text-blue-800">
                  <strong>Password requirements:</strong>
                </p>
                <ul className="mt-2 text-xs text-blue-700 list-disc list-inside space-y-1">
                  <li>At least 8 characters long</li>
                  <li>Contains both letters and numbers</li>
                  <li>Stronger passwords include uppercase, lowercase, numbers, and symbols</li>
                </ul>
              </div>
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
                  Resetting Password...
                </span>
              ) : (
                'Reset Password'
              )}
            </Button>
          </div>

          {/* Back Link */}
          <div className="text-center">
            <Link
              href="/reset-password"
              className="text-sm font-medium text-blue-600 hover:text-blue-500"
            >
              ← Start over
            </Link>
          </div>
        </form>
      </Card>
    </div>
  );
}
