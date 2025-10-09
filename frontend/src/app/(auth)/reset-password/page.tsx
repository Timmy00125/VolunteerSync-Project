'use client';

/**
 * Password Reset Request Page
 *
 * Initiates password reset by requesting user's email.
 * Returns security questions for verification.
 *
 * Flow:
 * 1. User enters email address
 * 2. API returns reset_token and security questions
 * 3. User redirected to verify page with reset_token
 *
 * Task: T105 (Part 1 of 3)
 * Requirements: FR-003 (Password Reset with Security Questions)
 * API: POST /api/v1/auth/password-reset/request
 */

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import Link from 'next/link';
import { passwordResetRequestSchema, type PasswordResetRequestData } from '@/lib/validations/auth';
import { post } from '@/lib/api/client';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card } from '@/components/ui/card';

interface PasswordResetResponse {
  reset_token: string;
  security_questions: string[];
}

export default function ResetPasswordPage() {
  const router = useRouter();
  const [isLoading, setIsLoading] = useState(false);
  const [apiError, setApiError] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<PasswordResetRequestData>({
    resolver: zodResolver(passwordResetRequestSchema),
    defaultValues: {
      email: '',
    },
  });

  // Handle form submission
  const onSubmit = async (data: PasswordResetRequestData) => {
    setIsLoading(true);
    setApiError(null);

    try {
      // Call password reset request API
      const response = await post<PasswordResetResponse>('/auth/password-reset/request', data, {
        skipAuth: true,
      });

      // Store reset token and questions in sessionStorage for next step
      // Using sessionStorage (not localStorage) for security - clears on browser close
      if (typeof window !== 'undefined') {
        sessionStorage.setItem('reset_token', response.reset_token);
        sessionStorage.setItem('security_questions', JSON.stringify(response.security_questions));
        sessionStorage.setItem('reset_email', data.email);
      }

      // Redirect to verification page
      router.push('/reset-password/verify');
    } catch (error: any) {
      console.error('Password reset request failed:', error);

      // Handle specific error codes
      if (error.status_code === 404) {
        // For security, don't reveal if email exists or not
        // Still show generic message
        setApiError(
          'If an account with that email exists, you will receive instructions to reset your password.'
        );
      } else if (error.status_code === 429) {
        setApiError('Too many password reset attempts. Please wait before trying again.');
      } else if (error.status_code === 400) {
        setApiError(error.message || 'Invalid email address. Please check and try again.');
      } else {
        setApiError('Unable to process password reset request. Please try again later.');
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
            <h1 className="text-3xl font-bold tracking-tight text-gray-900">Reset Password</h1>
            <p className="mt-2 text-sm text-gray-600">
              Enter your email address and we'll help you reset your password using security
              questions
            </p>
          </div>

          {/* API Error Display */}
          {apiError && (
            <div
              className={`rounded-md p-4 ${
                apiError.includes('If an account') ? 'bg-blue-50' : 'bg-red-50'
              }`}
            >
              <div className="flex">
                <div className="flex-shrink-0">
                  {apiError.includes('If an account') ? (
                    <svg className="h-5 w-5 text-blue-400" fill="currentColor" viewBox="0 0 20 20">
                      <path
                        fillRule="evenodd"
                        d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
                        clipRule="evenodd"
                      />
                    </svg>
                  ) : (
                    <svg className="h-5 w-5 text-red-400" fill="currentColor" viewBox="0 0 20 20">
                      <path
                        fillRule="evenodd"
                        d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                        clipRule="evenodd"
                      />
                    </svg>
                  )}
                </div>
                <div className="ml-3">
                  <p
                    className={`text-sm font-medium ${
                      apiError.includes('If an account') ? 'text-blue-800' : 'text-red-800'
                    }`}
                  >
                    {apiError}
                  </p>
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
                  Processing...
                </span>
              ) : (
                'Continue'
              )}
            </Button>
          </div>

          {/* Divider */}
          <div className="relative">
            <div className="absolute inset-0 flex items-center">
              <div className="w-full border-t border-gray-300" />
            </div>
            <div className="relative flex justify-center text-sm">
              <span className="bg-white px-2 text-gray-500">Or</span>
            </div>
          </div>

          {/* Back to Login */}
          <div className="text-center space-y-2">
            <Link href="/login" className="font-medium text-blue-600 hover:text-blue-500 block">
              Back to Login
            </Link>
            <Link href="/register" className="text-sm text-gray-600 hover:text-gray-500 block">
              Don't have an account? Create one
            </Link>
          </div>
        </form>
      </Card>
    </div>
  );
}
