'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { createOpportunitySchema } from '@/lib/validations';
import { useCreateOpportunity } from '@/lib/api';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Checkbox } from '@/components/ui/checkbox';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Loader2, Check, AlertCircle, MapPin, Users, Clock, FileText } from 'lucide-react';

/**
 * Create Opportunity Page (T117)
 *
 * Form for creating a new volunteer opportunity with:
 * - Basic details (title, description)
 * - Date and time
 * - Location (automatically geocoded)
 * - Capacity
 * - Requirements (skills, min age, documents)
 * - Recurrence options
 * - Publish immediately or save as draft
 */
export default function CreateOpportunityPage() {
  const router = useRouter();
  const createOpportunity = useCreateOpportunity();
  const [selectedCauses, setSelectedCauses] = useState<string[]>([]);
  const [selectedSkills, setSelectedSkills] = useState<string[]>([]);

  const {
    register: registerField,
    handleSubmit,
    watch,
    setValue,
    formState: { errors: formErrors, isSubmitting },
    setError,
  } = useForm({
    resolver: zodResolver(createOpportunitySchema) as any,
    defaultValues: {
      status: 'draft',
      is_recurring: false,
      recurrence_pattern: 'none',
      background_check_required: false,
      orientation_required: false,
    },
  });

  const register = registerField as any;
  const errors = formErrors as any;

  const isRecurring = watch('is_recurring');

  // Available options (should be fetched from API)
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

  const availableSkills = [
    { id: 'teaching', name: 'Teaching/Tutoring' },
    { id: 'construction', name: 'Construction' },
    { id: 'cooking', name: 'Cooking' },
    { id: 'first-aid', name: 'First Aid/CPR' },
    { id: 'gardening', name: 'Gardening' },
    { id: 'tech', name: 'Technology' },
    { id: 'event-planning', name: 'Event Planning' },
    { id: 'fundraising', name: 'Fundraising' },
  ];

  const onSubmit = async (data: any) => {
    try {
      // Add selected causes and skills to form data
      const formData = {
        ...data,
        cause_categories: selectedCauses,
        required_skills: selectedSkills,
      };

      const opportunity = await createOpportunity.mutateAsync(formData);

      // Success! Redirect based on status
      if (data.status === 'published') {
        router.push(`/opportunities/${opportunity.id}`);
      } else {
        router.push('/organization/opportunities');
      }
    } catch (error) {
      if (error instanceof Error) {
        setError('root', {
          type: 'manual',
          message: error.message || 'Failed to create opportunity. Please try again.',
        });
      }
    }
  };

  const toggleCause = (causeId: string) => {
    setSelectedCauses((prev) =>
      prev.includes(causeId) ? prev.filter((id) => id !== causeId) : [...prev, causeId]
    );
  };

  const toggleSkill = (skillId: string) => {
    setSelectedSkills((prev) =>
      prev.includes(skillId) ? prev.filter((id) => id !== skillId) : [...prev, skillId]
    );
  };

  return (
    <div className="mx-auto max-w-4xl space-y-8">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Create Volunteer Opportunity</h1>
        <p className="text-muted-foreground">
          Post a new volunteer opportunity to recruit volunteers for your event
        </p>
      </div>

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
        {/* Basic Information */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <FileText className="h-5 w-5" />
              Basic Information
            </CardTitle>
            <CardDescription>Provide details about the volunteer opportunity</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {/* Title */}
            <div className="space-y-2">
              <Label htmlFor="title">
                Title <span className="text-destructive">*</span>
              </Label>
              <Input
                id="title"
                placeholder="e.g., Community Garden Cleanup"
                {...register('title')}
                aria-invalid={!!errors.title}
              />
              {errors.title && <p className="text-sm text-destructive">{errors.title.message}</p>}
            </div>

            {/* Description */}
            <div className="space-y-2">
              <Label htmlFor="description">
                Description <span className="text-destructive">*</span>
              </Label>
              <textarea
                id="description"
                rows={6}
                className="flex w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                placeholder="Describe the volunteer opportunity, tasks involved, and what volunteers can expect"
                {...register('description')}
              />
              {errors.description && (
                <p className="text-sm text-destructive">{errors.description.message}</p>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Date and Time */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Clock className="h-5 w-5" />
              Date and Time
            </CardTitle>
            <CardDescription>When will this opportunity take place?</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid gap-4 md:grid-cols-2">
              {/* Start Date */}
              <div className="space-y-2">
                <Label htmlFor="start_date">
                  Start Date <span className="text-destructive">*</span>
                </Label>
                <Input id="start_date" type="date" {...register('start_date')} />
                {errors.start_date && (
                  <p className="text-sm text-destructive">{errors.start_date.message}</p>
                )}
              </div>

              {/* End Date */}
              <div className="space-y-2">
                <Label htmlFor="end_date">
                  End Date <span className="text-destructive">*</span>
                </Label>
                <Input id="end_date" type="date" {...register('end_date')} />
                {errors.end_date && (
                  <p className="text-sm text-destructive">{errors.end_date.message}</p>
                )}
              </div>
            </div>

            <div className="grid gap-4 md:grid-cols-2">
              {/* Start Time */}
              <div className="space-y-2">
                <Label htmlFor="start_time">
                  Start Time <span className="text-destructive">*</span>
                </Label>
                <Input
                  id="start_time"
                  type="time"
                  placeholder="09:00"
                  {...register('start_time')}
                />
                {errors.start_time && (
                  <p className="text-sm text-destructive">{errors.start_time.message}</p>
                )}
              </div>

              {/* End Time */}
              <div className="space-y-2">
                <Label htmlFor="end_time">
                  End Time <span className="text-destructive">*</span>
                </Label>
                <Input id="end_time" type="time" placeholder="17:00" {...register('end_time')} />
                {errors.end_time && (
                  <p className="text-sm text-destructive">{errors.end_time.message}</p>
                )}
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Location */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <MapPin className="h-5 w-5" />
              Location
            </CardTitle>
            <CardDescription>Address will be automatically geocoded</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="address_line_1">
                Street Address <span className="text-destructive">*</span>
              </Label>
              <Input
                id="address_line_1"
                placeholder="123 Main Street"
                {...register('address_line_1')}
              />
              {errors.address_line_1 && (
                <p className="text-sm text-destructive">{errors.address_line_1.message}</p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="address_line_2">Apartment, Suite, etc.</Label>
              <Input id="address_line_2" placeholder="Suite 100" {...register('address_line_2')} />
            </div>

            <div className="grid gap-4 md:grid-cols-3">
              <div className="space-y-2">
                <Label htmlFor="city">
                  City <span className="text-destructive">*</span>
                </Label>
                <Input id="city" placeholder="Portland" {...register('city')} />
                {errors.city && <p className="text-sm text-destructive">{errors.city.message}</p>}
              </div>

              <div className="space-y-2">
                <Label htmlFor="state">
                  State <span className="text-destructive">*</span>
                </Label>
                <Input id="state" placeholder="OR" {...register('state')} maxLength={2} />
                {errors.state && <p className="text-sm text-destructive">{errors.state.message}</p>}
              </div>

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

        {/* Capacity and Requirements */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Users className="h-5 w-5" />
              Capacity and Requirements
            </CardTitle>
            <CardDescription>Set volunteer limits and requirements</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid gap-4 md:grid-cols-2">
              {/* Capacity */}
              <div className="space-y-2">
                <Label htmlFor="capacity">
                  Volunteer Capacity <span className="text-destructive">*</span>
                </Label>
                <Input
                  id="capacity"
                  type="number"
                  min="1"
                  placeholder="20"
                  {...register('capacity', { valueAsNumber: true })}
                />
                {errors.capacity && (
                  <p className="text-sm text-destructive">{errors.capacity.message}</p>
                )}
              </div>

              {/* Minimum Age */}
              <div className="space-y-2">
                <Label htmlFor="minimum_age">Minimum Age</Label>
                <Input
                  id="minimum_age"
                  type="number"
                  min="0"
                  placeholder="18"
                  {...register('minimum_age', { valueAsNumber: true })}
                />
                {errors.minimum_age && (
                  <p className="text-sm text-destructive">{errors.minimum_age.message}</p>
                )}
              </div>
            </div>

            {/* Checkboxes */}
            <div className="space-y-3">
              <div className="flex items-center space-x-2">
                <Checkbox
                  id="background_check_required"
                  {...register('background_check_required')}
                />
                <Label
                  htmlFor="background_check_required"
                  className="text-sm font-normal leading-none"
                >
                  Background check required
                </Label>
              </div>

              <div className="flex items-center space-x-2">
                <Checkbox id="orientation_required" {...register('orientation_required')} />
                <Label htmlFor="orientation_required" className="text-sm font-normal leading-none">
                  Orientation/training required
                </Label>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Categories */}
        <Card>
          <CardHeader>
            <CardTitle>Cause Categories</CardTitle>
            <CardDescription>Select at least one category for this opportunity</CardDescription>
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
                  <Label htmlFor={cause.id} className="text-sm font-normal leading-none">
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

        {/* Required Skills */}
        <Card>
          <CardHeader>
            <CardTitle>Required Skills (Optional)</CardTitle>
            <CardDescription>
              Select any specific skills needed for this opportunity
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-3 md:grid-cols-2">
              {availableSkills.map((skill) => (
                <div key={skill.id} className="flex items-center space-x-2">
                  <Checkbox
                    id={skill.id}
                    checked={selectedSkills.includes(skill.id)}
                    onCheckedChange={() => toggleSkill(skill.id)}
                  />
                  <Label htmlFor={skill.id} className="text-sm font-normal leading-none">
                    {skill.name}
                  </Label>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* Recurrence Options */}
        <Card>
          <CardHeader>
            <CardTitle>Recurrence (Optional)</CardTitle>
            <CardDescription>Create recurring volunteer opportunities</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center space-x-2">
              <Checkbox
                id="is_recurring"
                checked={isRecurring}
                onCheckedChange={(checked) => setValue('is_recurring', checked as boolean)}
              />
              <Label htmlFor="is_recurring" className="text-sm font-normal leading-none">
                This is a recurring opportunity
              </Label>
            </div>

            {isRecurring && (
              <>
                <div className="space-y-2">
                  <Label htmlFor="recurrence_pattern">Recurrence Pattern</Label>
                  <Select
                    onValueChange={(value) => setValue('recurrence_pattern', value as any)}
                    defaultValue="none"
                  >
                    <SelectTrigger id="recurrence_pattern">
                      <SelectValue placeholder="Select pattern" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="none">None</SelectItem>
                      <SelectItem value="daily">Daily</SelectItem>
                      <SelectItem value="weekly">Weekly</SelectItem>
                      <SelectItem value="monthly">Monthly</SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                <div className="space-y-2">
                  <Label htmlFor="recurrence_end_date">Recurrence End Date</Label>
                  <Input
                    id="recurrence_end_date"
                    type="date"
                    {...register('recurrence_end_date')}
                  />
                </div>
              </>
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

        {/* Submit Buttons */}
        <div className="flex items-center gap-4">
          <Button
            type="submit"
            size="lg"
            disabled={isSubmitting || selectedCauses.length === 0}
            onClick={() => setValue('status', 'published')}
          >
            {isSubmitting ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Publishing...
              </>
            ) : (
              <>
                <Check className="mr-2 h-4 w-4" />
                Publish Now
              </>
            )}
          </Button>
          <Button
            type="submit"
            variant="outline"
            size="lg"
            disabled={isSubmitting || selectedCauses.length === 0}
            onClick={() => setValue('status', 'draft')}
          >
            Save as Draft
          </Button>
          <Button type="button" variant="ghost" onClick={() => router.back()}>
            Cancel
          </Button>
        </div>
      </form>
    </div>
  );
}
