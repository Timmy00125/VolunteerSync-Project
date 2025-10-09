'use client';

/**
 * Registration Page
 *
 * User registration form with:
 * - User type selection (volunteer, organization_admin)
 * - Email, password, name, phone
 * - 3 security questions for password reset
 * - Client-side validation with React Hook Form + Zod
 * - API integration to POST /auth/register
 *
 * Task: T103
 * Requirements: FR-002 (Registration), FR-003 (Security Questions)
 */

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import Link from 'next/link';
import { registerSchema, type RegisterFormData, SECURITY_QUESTIONS } from '@/lib/validations/auth';
import { post } from '@/lib/api/client';
import { useAuthStore } from '@/store/auth-store';
import type { AuthResponse } from '@/lib/api/types';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card } from '@/components/ui/card';

export default function RegisterPage() {
  const router = useRouter();
  const { login } = useAuthStore();
  const [isLoading, setIsLoading] = useState(false);
  const [apiError, setApiError] = useState<string | null>(null);
  const [step, setStep] = useState<'user-type' | 'details' | 'security'>('user-type');

  const {
    register,
    control,
    handleSubmit,
    watch,
    formState: { errors },
    setValue,
  } = useForm<RegisterFormData>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      email: '',
      password: '',
      first_name: '',
      last_name: '',
      phone: '',
      user_type: 'volunteer',
      security_questions: [
        { question: '', answer: '' },
        { question: '', answer: '' },
        { question: '', answer: '' },
      ],
    },
  });

  const userType = watch('user_type');
  const securityQuestions = watch('security_questions');

  // Handle user type selection
  const handleUserTypeSelect = (type: 'volunteer' | 'organization_admin') => {
    setValue('user_type', type);
    setStep('details');
  };

  // Handle form submission
  const onSubmit = async (data: RegisterFormData) => {
    setIsLoading(true);
    setApiError(null);

    try {
      // Validate security questions are unique
      const questionTexts = data.security_questions.map((q) => q.question);
      const uniqueQuestions = new Set(questionTexts);
      if (uniqueQuestions.size !== questionTexts.length) {
        setApiError('Please select different security questions for each slot');
        setIsLoading(false);
        return;
      }

      // Call registration API
      const response = await post<AuthResponse>('/auth/register', data, {
        skipAuth: true,
      });

      // Store user and tokens in auth store
      login(response.user, response.tokens);

      // Redirect based on user type
      if (data.user_type === 'volunteer') {
        router.push('/volunteer/profile'); // Complete volunteer profile
      } else {
        router.push('/organization/new'); // Create organization
      }
    } catch (error: any) {
      console.error('Registration failed:', error);

      // Handle specific error codes
      if (error.status_code === 409) {
        setApiError('An account with this email already exists. Please log in instead.');
      } else if (error.status_code === 429) {
        setApiError('Too many registration attempts. Please try again later.');
      } else if (error.status_code === 400) {
        setApiError(error.message || 'Invalid data. Please check your inputs.');
      } else {
        setApiError('Registration failed. Please try again.');
      }
    } finally {
      setIsLoading(false);
    }
  };

  // Render user type selection
  if (step === 'user-type') {
    return (
      <div className="flex min-h-screen items-center justify-center bg-gray-50 px-4 py-12 sm:px-6 lg:px-8">
        <div className="w-full max-w-2xl space-y-8">
          <div className="text-center">
            <h1 className="text-4xl font-bold tracking-tight text-gray-900">Join VolunteerSync</h1>
            <p className="mt-2 text-lg text-gray-600">Choose how you'd like to get started</p>
          </div>

          <div className="grid grid-cols-1 gap-6 sm:grid-cols-2">
            {/* Volunteer Card */}
            <Card
              className="cursor-pointer border-2 border-gray-200 p-6 transition-all hover:border-blue-500 hover:shadow-lg"
              onClick={() => handleUserTypeSelect('volunteer')}
            >
              <div className="flex flex-col items-center text-center">
                <div className="mb-4 flex h-20 w-20 items-center justify-center rounded-full bg-blue-100">
                  <svg
                    className="h-10 w-10 text-blue-600"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"
                    />
                  </svg>
                </div>
                <h2 className="text-xl font-semibold text-gray-900">I'm a Volunteer</h2>
                <p className="mt-2 text-sm text-gray-600">
                  Find opportunities to make a difference in your community
                </p>
                <ul className="mt-4 space-y-2 text-left text-sm text-gray-700">
                  <li className="flex items-start">
                    <span className="mr-2">✓</span>
                    <span>Discover local volunteer opportunities</span>
                  </li>
                  <li className="flex items-start">
                    <span className="mr-2">✓</span>
                    <span>Track your volunteer hours</span>
                  </li>
                  <li className="flex items-start">
                    <span className="mr-2">✓</span>
                    <span>Earn achievement badges</span>
                  </li>
                </ul>
              </div>
            </Card>

            {/* Organization Admin Card */}
            <Card
              className="cursor-pointer border-2 border-gray-200 p-6 transition-all hover:border-green-500 hover:shadow-lg"
              onClick={() => handleUserTypeSelect('organization_admin')}
            >
              <div className="flex flex-col items-center text-center">
                <div className="mb-4 flex h-20 w-20 items-center justify-center rounded-full bg-green-100">
                  <svg
                    className="h-10 w-10 text-green-600"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"
                    />
                  </svg>
                </div>
                <h2 className="text-xl font-semibold text-gray-900">I'm an Organization</h2>
                <p className="mt-2 text-sm text-gray-600">
                  Connect with volunteers and manage your programs
                </p>
                <ul className="mt-4 space-y-2 text-left text-sm text-gray-700">
                  <li className="flex items-start">
                    <span className="mr-2">✓</span>
                    <span>Post volunteer opportunities</span>
                  </li>
                  <li className="flex items-start">
                    <span className="mr-2">✓</span>
                    <span>Manage volunteer registrations</span>
                  </li>
                  <li className="flex items-start">
                    <span className="mr-2">✓</span>
                    <span>Track volunteer impact</span>
                  </li>
                </ul>
              </div>
            </Card>
          </div>

          <div className="text-center text-sm text-gray-600">
            Already have an account?{' '}
            <Link href="/login" className="font-medium text-blue-600 hover:text-blue-500">
              Sign in
            </Link>
          </div>
        </div>
      </div>
    );
  }

  // Render registration form
  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50 px-4 py-12 sm:px-6 lg:px-8">
      <Card className="w-full max-w-2xl p-8">
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
          {/* Header */}
          <div className="text-center">
            <h1 className="text-3xl font-bold tracking-tight text-gray-900">Create Your Account</h1>
            <p className="mt-2 text-sm text-gray-600">
              Register as a{' '}
              <span className="font-medium">
                {userType === 'volunteer' ? 'Volunteer' : 'Organization Admin'}
              </span>
            </p>
            <Button
              type="button"
              variant="outline"
              size="sm"
              className="mt-2"
              onClick={() => setStep('user-type')}
            >
              Change user type
            </Button>
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

          {/* Step 1: Basic Details */}
          {step === 'details' && (
            <div className="space-y-4">
              <h2 className="text-lg font-semibold text-gray-900">Basic Information</h2>

              {/* Email */}
              <div>
                <Label htmlFor="email">Email Address *</Label>
                <Input
                  id="email"
                  type="email"
                  autoComplete="email"
                  {...register('email')}
                  className={errors.email ? 'border-red-500' : ''}
                />
                {errors.email && (
                  <p className="mt-1 text-sm text-red-600">{errors.email.message}</p>
                )}
              </div>

              {/* Password */}
              <div>
                <Label htmlFor="password">Password *</Label>
                <Input
                  id="password"
                  type="password"
                  autoComplete="new-password"
                  {...register('password')}
                  className={errors.password ? 'border-red-500' : ''}
                />
                {errors.password && (
                  <p className="mt-1 text-sm text-red-600">{errors.password.message}</p>
                )}
                <p className="mt-1 text-xs text-gray-500">
                  Minimum 8 characters, must include letters and numbers
                </p>
              </div>

              {/* First Name */}
              <div>
                <Label htmlFor="first_name">First Name *</Label>
                <Input
                  id="first_name"
                  type="text"
                  autoComplete="given-name"
                  {...register('first_name')}
                  className={errors.first_name ? 'border-red-500' : ''}
                />
                {errors.first_name && (
                  <p className="mt-1 text-sm text-red-600">{errors.first_name.message}</p>
                )}
              </div>

              {/* Last Name */}
              <div>
                <Label htmlFor="last_name">Last Name *</Label>
                <Input
                  id="last_name"
                  type="text"
                  autoComplete="family-name"
                  {...register('last_name')}
                  className={errors.last_name ? 'border-red-500' : ''}
                />
                {errors.last_name && (
                  <p className="mt-1 text-sm text-red-600">{errors.last_name.message}</p>
                )}
              </div>

              {/* Phone (Optional) */}
              <div>
                <Label htmlFor="phone">Phone Number (Optional)</Label>
                <Input
                  id="phone"
                  type="tel"
                  autoComplete="tel"
                  placeholder="+1234567890"
                  {...register('phone')}
                  className={errors.phone ? 'border-red-500' : ''}
                />
                {errors.phone && (
                  <p className="mt-1 text-sm text-red-600">{errors.phone.message}</p>
                )}
              </div>

              <Button type="button" onClick={() => setStep('security')} className="w-full">
                Next: Security Questions
              </Button>
            </div>
          )}

          {/* Step 2: Security Questions */}
          {step === 'security' && (
            <div className="space-y-4">
              <div>
                <h2 className="text-lg font-semibold text-gray-900">Security Questions</h2>
                <p className="mt-1 text-sm text-gray-600">
                  Choose 3 security questions for password recovery. You'll need to answer 2 out of
                  3 correctly to reset your password.
                </p>
              </div>

              {[0, 1, 2].map((index) => (
                <div key={index} className="space-y-2 rounded-lg border border-gray-200 p-4">
                  <Label htmlFor={`security_question_${index}`}>
                    Security Question {index + 1} *
                  </Label>

                  {/* Question Dropdown */}
                  <Controller
                    name={`security_questions.${index}.question`}
                    control={control}
                    render={({ field }) => (
                      <select
                        {...field}
                        id={`security_question_${index}`}
                        className={`w-full rounded-md border px-3 py-2 ${
                          errors.security_questions?.[index]?.question
                            ? 'border-red-500'
                            : 'border-gray-300'
                        }`}
                      >
                        <option value="">-- Select a question --</option>
                        {SECURITY_QUESTIONS.map((question) => (
                          <option
                            key={question}
                            value={question}
                            disabled={securityQuestions.some(
                              (sq, i) => i !== index && sq.question === question
                            )}
                          >
                            {question}
                          </option>
                        ))}
                      </select>
                    )}
                  />
                  {errors.security_questions?.[index]?.question && (
                    <p className="text-sm text-red-600">
                      {errors.security_questions[index]?.question?.message}
                    </p>
                  )}

                  {/* Answer Input */}
                  <Label htmlFor={`security_answer_${index}`}>Your Answer *</Label>
                  <Input
                    id={`security_answer_${index}`}
                    type="text"
                    placeholder="Enter your answer"
                    {...register(`security_questions.${index}.answer`)}
                    className={errors.security_questions?.[index]?.answer ? 'border-red-500' : ''}
                  />
                  {errors.security_questions?.[index]?.answer && (
                    <p className="text-sm text-red-600">
                      {errors.security_questions[index]?.answer?.message}
                    </p>
                  )}
                </div>
              ))}

              <div className="flex space-x-4">
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setStep('details')}
                  className="flex-1"
                >
                  Back
                </Button>
                <Button type="submit" disabled={isLoading} className="flex-1">
                  {isLoading ? 'Creating Account...' : 'Create Account'}
                </Button>
              </div>
            </div>
          )}

          {/* Footer */}
          <div className="text-center text-sm text-gray-600">
            Already have an account?{' '}
            <Link href="/login" className="font-medium text-blue-600 hover:text-blue-500">
              Sign in
            </Link>
          </div>
        </form>
      </Card>
    </div>
  );
}
