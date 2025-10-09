'use client';

/**
 * Password Reset Verify Page
 *
 * Verify identity by answering security questions.
 * Requires 2 of 3 correct answers to proceed.
 *
 * Flow:
 * 1. Display 3 security questions from previous step
 * 2. User answers all 3 questions
 * 3. API verifies answers (2 of 3 must be correct)
 * 4. On success, receive verified_token and redirect to confirm page
 *
 * Task: T105 (Part 2 of 3)
 * Requirements: FR-003a (2 of 3 security questions correct)
 * API: POST /api/v1/auth/password-reset/verify
 */

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import Link from 'next/link';
import { passwordResetVerifySchema, type PasswordResetVerifyData } from '@/lib/validations/auth';
import { post } from '@/lib/api/client';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card } from '@/components/ui/card';

interface PasswordResetVerifyResponse {
  verified_token: string;
}

export default function ResetPasswordVerifyPage() {
  const router = useRouter();
  const [isLoading, setIsLoading] = useState(false);
  const [apiError, setApiError] = useState<string | null>(null);
  const [securityQuestions, setSecurityQuestions] = useState<string[]>([]);
  const [email, setEmail] = useState<string>('');

  const {
    register,
    handleSubmit,
    formState: { errors },
    setValue,
  } = useForm<PasswordResetVerifyData>({
    resolver: zodResolver(passwordResetVerifySchema),
    defaultValues: {
      reset_token: '',
      answers: ['', '', ''],
    },
  });

  // Load data from sessionStorage on mount
  useEffect(() => {
    if (typeof window === 'undefined') return;

    const token = sessionStorage.getItem('reset_token');
    const questionsJson = sessionStorage.getItem('security_questions');
    const storedEmail = sessionStorage.getItem('reset_email');

    if (!token || !questionsJson) {
      // No reset token found - redirect to request page
      router.push('/reset-password');
      return;
    }

    try {
      const questions = JSON.parse(questionsJson);
      setSecurityQuestions(questions);
      setEmail(storedEmail || '');
      setValue('reset_token', token);
    } catch (error) {
      console.error('Failed to parse security questions:', error);
      router.push('/reset-password');
    }
  }, [router, setValue]);

  // Handle form submission
  const onSubmit = async (data: PasswordResetVerifyData) => {
    setIsLoading(true);
    setApiError(null);

    try {
      // Call password reset verify API
      const response = await post<PasswordResetVerifyResponse>(
        '/auth/password-reset/verify',
        data,
        { skipAuth: true }
      );

      // Store verified token for final step
      if (typeof window !== 'undefined') {
        sessionStorage.setItem('verified_token', response.verified_token);
        // Clean up reset token as it's no longer needed
        sessionStorage.removeItem('reset_token');
        sessionStorage.removeItem('security_questions');
      }

      // Redirect to confirm password page
      router.push('/reset-password/confirm');
    } catch (error: any) {
      console.error('Security question verification failed:', error);

      // Handle specific error codes
      if (error.status_code === 401) {
        setApiError(
          'Incorrect answers. You must answer at least 2 out of 3 security questions correctly. Please try again.'
        );
      } else if (error.status_code === 400) {
        setApiError(error.message || 'Invalid request. Please try again.');
      } else if (error.status_code === 429) {
        setApiError('Too many verification attempts. Please wait before trying again.');
      } else if (error.status_code === 404) {
        setApiError('Password reset session expired. Please start over from the beginning.');
        // Clear session storage and redirect after delay
        setTimeout(() => {
          if (typeof window !== 'undefined') {
            sessionStorage.removeItem('reset_token');
            sessionStorage.removeItem('security_questions');
            sessionStorage.removeItem('reset_email');
          }
          router.push('/reset-password');
        }, 3000);
      } else {
        setApiError('Unable to verify answers. Please try again later.');
      }
    } finally {
      setIsLoading(false);
    }
  };

  // Show loading while checking for reset token
  if (securityQuestions.length === 0) {
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
          <p className="mt-4 text-sm text-gray-600">Loading security questions...</p>
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
            <h1 className="text-3xl font-bold tracking-tight text-gray-900">
              Verify Your Identity
            </h1>
            {email && (
              <p className="mt-2 text-sm text-gray-600">
                Resetting password for: <span className="font-medium">{email}</span>
              </p>
            )}
            <p className="mt-2 text-sm text-gray-600">
              Answer your security questions to continue. You must answer at least 2 out of 3
              questions correctly.
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

          {/* Security Questions */}
          <div className="space-y-4">
            {securityQuestions.map((question, index) => (
              <div key={index}>
                <Label htmlFor={`answer-${index}`}>
                  Question {index + 1}: {question}
                </Label>
                <Input
                  id={`answer-${index}`}
                  type="text"
                  autoComplete="off"
                  placeholder="Your answer"
                  {...register(`answers.${index}` as const)}
                  className={errors.answers?.[index] ? 'border-red-500' : ''}
                  disabled={isLoading}
                />
                {errors.answers?.[index] && (
                  <p className="mt-1 text-sm text-red-600">{errors.answers[index]?.message}</p>
                )}
              </div>
            ))}
          </div>

          {/* Info Box */}
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
                  <strong>Tip:</strong> Answers are case-insensitive. You need to get at least 2 out
                  of 3 questions correct to proceed.
                </p>
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
                  Verifying...
                </span>
              ) : (
                'Verify Answers'
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
