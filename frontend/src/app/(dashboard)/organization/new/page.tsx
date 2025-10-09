'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { createOrganizationSchema } from '@/lib/validations';
import { useCreateOrganization } from '@/lib/api';
import type { CreateOrganizationInput } from '@/lib/api/types';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Checkbox } from '@/components/ui/checkbox';
import { Building2, Loader2, Check, AlertCircle } from 'lucide-react';

/**
 * Create Organization Page (T116)
 *
 * Form for creating a new organization profile with:
 * - Organization details (name, mission, description)
 * - Contact information (email, phone, website)
 * - Address (automatically geocoded on submit)
 * - Logo and banner images
 * - Cause categories
 *
 * Organizations are auto-verified on creation (FR-015)
 */
export default function CreateOrganizationPage() {
  const router = useRouter();
  const createOrganization = useCreateOrganization();
  const [selectedCauses, setSelectedCauses] = useState<string[]>([]);

  const {
    register: registerField,
    handleSubmit,
    formState: { errors: formErrors, isSubmitting },
    setError,
  } = useForm({
    resolver: zodResolver(createOrganizationSchema) as any,
  });

  const register = registerField as any;
  const errors = formErrors as any;

  // Available cause categories (should be fetched from API)
  const availableCauses = [
    { id: 'education', name: 'Education' },
    { id: 'environment', name: 'Environment' },
    { id: 'health', name: 'Health & Wellness' },
    { id: 'animals', name: 'Animals' },
    { id: 'community', name: 'Community Development' },
    { id: 'arts', name: 'Arts & Culture' },
    { id: 'hunger', name: 'Hunger & Homelessness' },
    { id: 'advocacy', name: 'Advocacy & Human Rights' },
  ];

  const onSubmit = async (data: any) => {
    try {
      // Transform form data to match API expectations
      const apiData: CreateOrganizationInput = {
        name: data.name,
        description: data.description || '',
        mission: data.mission_statement,
        website: data.website || undefined,
        email: data.email,
        phone: data.phone || undefined,
        address: data.address_line_1 + (data.address_line_2 ? `, ${data.address_line_2}` : ''),
        city: data.city,
        state: data.state,
        zip_code: data.postal_code,
        country: data.country,
        logo_url: data.logo_url || undefined,
        cause_ids: selectedCauses,
      };

      await createOrganization.mutateAsync(apiData);

      // Success! Redirect to organization dashboard
      router.push('/organization');
    } catch (error) {
      // Handle error
      if (error instanceof Error) {
        setError('root', {
          type: 'manual',
          message: error.message || 'Failed to create organization. Please try again.',
        });
      }
    }
  };

  const toggleCause = (causeId: string) => {
    setSelectedCauses((prev) =>
      prev.includes(causeId) ? prev.filter((id) => id !== causeId) : [...prev, causeId]
    );
  };

  return (
    <div className="mx-auto max-w-4xl space-y-8">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Create Organization</h1>
        <p className="text-muted-foreground">
          Set up your organization profile to start posting volunteer opportunities
        </p>
      </div>

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
        {/* Basic Information */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Building2 className="h-5 w-5" />
              Basic Information
            </CardTitle>
            <CardDescription>
              Tell volunteers about your organization and its mission
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {/* Organization Name */}
            <div className="space-y-2">
              <Label htmlFor="name">
                Organization Name <span className="text-destructive">*</span>
              </Label>
              <Input
                id="name"
                placeholder="e.g., Community Food Bank"
                {...register('name')}
                aria-invalid={!!errors.name}
              />
              {errors.name && <p className="text-sm text-destructive">{errors.name.message}</p>}
            </div>

            {/* Mission Statement */}
            <div className="space-y-2">
              <Label htmlFor="mission_statement">Mission Statement</Label>
              <Input
                id="mission_statement"
                placeholder="Brief statement of your organization's purpose"
                {...register('mission_statement')}
              />
              {errors.mission_statement && (
                <p className="text-sm text-destructive">{errors.mission_statement.message}</p>
              )}
            </div>

            {/* Description */}
            <div className="space-y-2">
              <Label htmlFor="description">Description</Label>
              <textarea
                id="description"
                rows={4}
                className="flex w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                placeholder="Detailed description of your organization and its activities"
                {...register('description')}
              />
              {errors.description && (
                <p className="text-sm text-destructive">{errors.description.message}</p>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Contact Information */}
        <Card>
          <CardHeader>
            <CardTitle>Contact Information</CardTitle>
            <CardDescription>How volunteers can reach your organization</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid gap-4 md:grid-cols-2">
              {/* Email */}
              <div className="space-y-2">
                <Label htmlFor="email">
                  Email <span className="text-destructive">*</span>
                </Label>
                <Input
                  id="email"
                  type="email"
                  placeholder="contact@example.org"
                  {...register('email')}
                  aria-invalid={!!errors.email}
                />
                {errors.email && <p className="text-sm text-destructive">{errors.email.message}</p>}
              </div>

              {/* Phone */}
              <div className="space-y-2">
                <Label htmlFor="phone">
                  Phone <span className="text-destructive">*</span>
                </Label>
                <Input
                  id="phone"
                  type="tel"
                  placeholder="(555) 123-4567"
                  {...register('phone')}
                  aria-invalid={!!errors.phone}
                />
                {errors.phone && <p className="text-sm text-destructive">{errors.phone.message}</p>}
              </div>
            </div>

            {/* Website */}
            <div className="space-y-2">
              <Label htmlFor="website">
                Website <span className="text-destructive">*</span>
              </Label>
              <Input
                id="website"
                type="url"
                placeholder="https://example.org"
                {...register('website')}
                aria-invalid={!!errors.website}
              />
              {errors.website && (
                <p className="text-sm text-destructive">{errors.website.message}</p>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Address */}
        <Card>
          <CardHeader>
            <CardTitle>Address</CardTitle>
            <CardDescription>Physical location will be automatically geocoded</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {/* Address Line 1 */}
            <div className="space-y-2">
              <Label htmlFor="address_line_1">
                Street Address <span className="text-destructive">*</span>
              </Label>
              <Input
                id="address_line_1"
                placeholder="123 Main Street"
                {...register('address_line_1')}
                aria-invalid={!!errors.address_line_1}
              />
              {errors.address_line_1 && (
                <p className="text-sm text-destructive">{errors.address_line_1.message}</p>
              )}
            </div>

            {/* Address Line 2 */}
            <div className="space-y-2">
              <Label htmlFor="address_line_2">Apartment, Suite, etc.</Label>
              <Input id="address_line_2" placeholder="Suite 100" {...register('address_line_2')} />
            </div>

            <div className="grid gap-4 md:grid-cols-3">
              {/* City */}
              <div className="space-y-2">
                <Label htmlFor="city">
                  City <span className="text-destructive">*</span>
                </Label>
                <Input id="city" placeholder="Portland" {...register('city')} />
                {errors.city && <p className="text-sm text-destructive">{errors.city.message}</p>}
              </div>

              {/* State */}
              <div className="space-y-2">
                <Label htmlFor="state">
                  State <span className="text-destructive">*</span>
                </Label>
                <Input id="state" placeholder="OR" {...register('state')} maxLength={2} />
                {errors.state && <p className="text-sm text-destructive">{errors.state.message}</p>}
              </div>

              {/* Postal Code */}
              <div className="space-y-2">
                <Label htmlFor="postal_code">
                  ZIP Code <span className="text-destructive">*</span>
                </Label>
                <Input id="postal_code" placeholder="97201" {...register('postal_code')} />
                {errors.postal_code && (
                  <p className="text-sm text-destructive">{errors.postal_code.message}</p>
                )}
              </div>
            </div>

            {/* Country */}
            <div className="space-y-2">
              <Label htmlFor="country">
                Country <span className="text-destructive">*</span>
              </Label>
              <Input id="country" placeholder="United States" {...register('country')} />
              {errors.country && (
                <p className="text-sm text-destructive">{errors.country.message}</p>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Branding */}
        <Card>
          <CardHeader>
            <CardTitle>Branding</CardTitle>
            <CardDescription>Upload your organization&apos;s visual identity</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {/* Logo URL */}
            <div className="space-y-2">
              <Label htmlFor="logo_url">
                Logo URL <span className="text-destructive">*</span>
              </Label>
              <Input
                id="logo_url"
                type="url"
                placeholder="https://example.org/logo.png"
                {...register('logo_url')}
              />
              {errors.logo_url && (
                <p className="text-sm text-destructive">{errors.logo_url.message}</p>
              )}
              <p className="text-xs text-muted-foreground">
                Recommended: Square image, at least 200x200px
              </p>
            </div>

            {/* Banner URL */}
            <div className="space-y-2">
              <Label htmlFor="banner_url">
                Banner URL <span className="text-destructive">*</span>
              </Label>
              <Input
                id="banner_url"
                type="url"
                placeholder="https://example.org/banner.png"
                {...register('banner_url')}
              />
              {errors.banner_url && (
                <p className="text-sm text-destructive">{errors.banner_url.message}</p>
              )}
              <p className="text-xs text-muted-foreground">
                Recommended: Wide image, at least 1200x400px
              </p>
            </div>
          </CardContent>
        </Card>

        {/* Cause Categories */}
        <Card>
          <CardHeader>
            <CardTitle>Cause Categories</CardTitle>
            <CardDescription>
              Select at least one category that best describes your organization&apos;s focus
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-3 md:grid-cols-2">
              {availableCauses.map((cause) => (
                <div key={cause.id} className="flex items-center space-x-2">
                  <Checkbox
                    id={cause.id}
                    checked={selectedCauses.includes(cause.id)}
                    onCheckedChange={() => toggleCause(cause.id)}
                  />
                  <Label
                    htmlFor={cause.id}
                    className="text-sm font-normal leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                  >
                    {cause.name}
                  </Label>
                </div>
              ))}
            </div>
            {selectedCauses.length === 0 && (
              <p className="mt-2 text-sm text-destructive">
                Please select at least one cause category
              </p>
            )}
          </CardContent>
        </Card>

        {/* Form Error */}
        {errors.root && (
          <Card className="border-destructive">
            <CardContent className="flex items-center gap-2 py-4">
              <AlertCircle className="h-5 w-5 text-destructive" />
              <p className="text-sm text-destructive">{errors.root.message}</p>
            </CardContent>
          </Card>
        )}

        {/* Submit Button */}
        <div className="flex items-center gap-4">
          <Button type="submit" size="lg" disabled={isSubmitting || selectedCauses.length === 0}>
            {isSubmitting ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Creating Organization...
              </>
            ) : (
              <>
                <Check className="mr-2 h-4 w-4" />
                Create Organization
              </>
            )}
          </Button>
          <Button type="button" variant="outline" onClick={() => router.back()}>
            Cancel
          </Button>
        </div>
      </form>
    </div>
  );
}
