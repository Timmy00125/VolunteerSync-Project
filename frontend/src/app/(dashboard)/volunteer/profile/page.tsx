'use client';

import { useState } from 'react';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Checkbox } from '@/components/ui/checkbox';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { useVolunteerProfile, useUpdateVolunteerProfile } from '@/lib/api';
import { Loader2, Save, User, MapPin, Calendar, Shield, Bell, Heart, Award } from 'lucide-react';

// Validation schema for volunteer profile
const profileSchema = z.object({
  profile_photo_url: z.string().url().optional().or(z.literal('')),
  biography: z.string().max(1000).optional().or(z.literal('')),
  location: z.string().max(255).optional().or(z.literal('')),
  preferred_time: z.enum(['morning', 'afternoon', 'evening', 'flexible']).optional(),
  emergency_contact_name: z.string().max(200).optional().or(z.literal('')),
  emergency_contact_phone: z.string().max(20).optional().or(z.literal('')),
  availability: z.object({
    monday: z.boolean(),
    tuesday: z.boolean(),
    wednesday: z.boolean(),
    thursday: z.boolean(),
    friday: z.boolean(),
    saturday: z.boolean(),
    sunday: z.boolean(),
  }),
  privacy_settings: z.object({
    show_hours: z.boolean(),
    show_events: z.boolean(),
    show_organizations: z.boolean(),
  }),
  notification_settings: z.object({
    in_app: z.boolean(),
    browser_push: z.boolean(),
  }),
  skill_ids: z.array(z.string()).optional(),
  interest_ids: z.array(z.string()).optional(),
});

type ProfileFormData = z.infer<typeof profileSchema>;

/**
 * Volunteer Profile Page
 *
 * Allows volunteers to manage their profile including:
 * - Bio, location, photo
 * - Availability (days of week)
 * - Preferred time of day
 * - Skills and interests
 * - Privacy settings
 * - Emergency contact
 * - Notification preferences
 */
export default function VolunteerProfilePage() {
  const { data: profile, isLoading, error } = useVolunteerProfile();
  const updateProfile = useUpdateVolunteerProfile();
  const [isSaving, setIsSaving] = useState(false);

  const {
    register,
    control,
    handleSubmit,
    formState: { errors, isDirty },
    reset,
  } = useForm<ProfileFormData>({
    resolver: zodResolver(profileSchema),
    values: profile
      ? {
          profile_photo_url: profile.profile_photo_url || '',
          biography: profile.biography || '',
          location: profile.location || '',
          preferred_time: profile.preferred_time,
          emergency_contact_name: profile.emergency_contact_name || '',
          emergency_contact_phone: profile.emergency_contact_phone || '',
          availability: {
            monday: profile.availability_monday,
            tuesday: profile.availability_tuesday,
            wednesday: profile.availability_wednesday,
            thursday: profile.availability_thursday,
            friday: profile.availability_friday,
            saturday: profile.availability_saturday,
            sunday: profile.availability_sunday,
          },
          privacy_settings: {
            show_hours: profile.privacy_show_hours,
            show_events: profile.privacy_show_events,
            show_organizations: profile.privacy_show_organizations,
          },
          notification_settings: {
            in_app: profile.notification_in_app,
            browser_push: profile.notification_browser_push,
          },
          skill_ids: profile.skills?.map((s) => s.id) || [],
          interest_ids: profile.interests?.map((i) => i.id) || [],
        }
      : undefined,
  });

  const onSubmit = async (data: ProfileFormData) => {
    setIsSaving(true);
    try {
      await updateProfile.mutateAsync(data);
      reset(data); // Reset form to mark as not dirty
    } catch (error) {
      console.error('Failed to update profile:', error);
    } finally {
      setIsSaving(false);
    }
  };

  // Loading state
  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="space-y-2">
          <div className="h-9 w-64 animate-pulse rounded-md bg-muted" />
          <div className="h-5 w-96 animate-pulse rounded-md bg-muted" />
        </div>
        <div className="grid gap-6">
          {[...Array(3)].map((_, i) => (
            <Card key={i}>
              <CardHeader>
                <div className="h-6 w-32 animate-pulse rounded-md bg-muted" />
              </CardHeader>
              <CardContent>
                <div className="h-48 animate-pulse rounded-md bg-muted" />
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    );
  }

  // Error state
  if (error) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Profile</h1>
        </div>
        <Card className="border-destructive">
          <CardHeader>
            <CardTitle className="text-destructive">Error Loading Profile</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">
              {error instanceof Error
                ? error.message
                : 'Failed to load profile data. Please try again later.'}
            </p>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Profile</h1>
          <p className="text-muted-foreground">
            Manage your volunteer profile, availability, and preferences
          </p>
        </div>
        <Button onClick={handleSubmit(onSubmit)} disabled={!isDirty || isSaving} className="gap-2">
          {isSaving ? (
            <>
              <Loader2 className="h-4 w-4 animate-spin" />
              Saving...
            </>
          ) : (
            <>
              <Save className="h-4 w-4" />
              Save Changes
            </>
          )}
        </Button>
      </div>

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
        {/* Basic Information */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <User className="h-5 w-5" />
              Basic Information
            </CardTitle>
            <CardDescription>
              Your public profile information visible to organizations
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="profile_photo_url">Profile Photo URL</Label>
              <Input
                id="profile_photo_url"
                {...register('profile_photo_url')}
                placeholder="https://example.com/photo.jpg"
                type="url"
              />
              {errors.profile_photo_url && (
                <p className="text-sm text-destructive">{errors.profile_photo_url.message}</p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="biography">Biography</Label>
              <Textarea
                id="biography"
                {...register('biography')}
                placeholder="Tell organizations about yourself, your interests, and why you volunteer..."
                rows={5}
              />
              <p className="text-xs text-muted-foreground">
                {profile?.biography?.length || 0}/1000 characters
              </p>
              {errors.biography && (
                <p className="text-sm text-destructive">{errors.biography.message}</p>
              )}
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
            <CardDescription>
              Help organizations find volunteer opportunities near you
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="location">Location</Label>
              <Input
                id="location"
                {...register('location')}
                placeholder="City, State or full address"
              />
              <p className="text-xs text-muted-foreground">
                This will be geocoded to find opportunities near you
              </p>
              {errors.location && (
                <p className="text-sm text-destructive">{errors.location.message}</p>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Availability */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Calendar className="h-5 w-5" />
              Availability
            </CardTitle>
            <CardDescription>
              Select which days you're typically available to volunteer
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-3">
              {['monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday'].map(
                (day) => (
                  <div key={day} className="flex items-center space-x-2">
                    <Controller
                      name={`availability.${day as keyof ProfileFormData['availability']}`}
                      control={control}
                      render={({ field }) => (
                        <Checkbox id={day} checked={field.value} onCheckedChange={field.onChange} />
                      )}
                    />
                    <Label htmlFor={day} className="cursor-pointer capitalize">
                      {day}
                    </Label>
                  </div>
                )
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="preferred_time">Preferred Time of Day</Label>
              <Controller
                name="preferred_time"
                control={control}
                render={({ field }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger>
                      <SelectValue placeholder="Select preferred time" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="morning">Morning</SelectItem>
                      <SelectItem value="afternoon">Afternoon</SelectItem>
                      <SelectItem value="evening">Evening</SelectItem>
                      <SelectItem value="flexible">Flexible</SelectItem>
                    </SelectContent>
                  </Select>
                )}
              />
            </div>
          </CardContent>
        </Card>

        {/* Skills & Interests (Placeholder - will be enhanced with actual selectors) */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Award className="h-5 w-5" />
              Skills & Interests
            </CardTitle>
            <CardDescription>
              Select your skills and causes you're interested in supporting
            </CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">
              Skills and interests management will be enhanced with multi-select functionality in a
              future update.
            </p>
            {/* TODO: Add multi-select for skills and interests */}
          </CardContent>
        </Card>

        {/* Emergency Contact */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Heart className="h-5 w-5" />
              Emergency Contact
            </CardTitle>
            <CardDescription>
              Contact information in case of emergencies during events
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="emergency_contact_name">Name</Label>
              <Input
                id="emergency_contact_name"
                {...register('emergency_contact_name')}
                placeholder="Emergency contact name"
              />
              {errors.emergency_contact_name && (
                <p className="text-sm text-destructive">{errors.emergency_contact_name.message}</p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="emergency_contact_phone">Phone Number</Label>
              <Input
                id="emergency_contact_phone"
                {...register('emergency_contact_phone')}
                placeholder="+1 (555) 123-4567"
                type="tel"
              />
              {errors.emergency_contact_phone && (
                <p className="text-sm text-destructive">{errors.emergency_contact_phone.message}</p>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Privacy Settings */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Shield className="h-5 w-5" />
              Privacy Settings
            </CardTitle>
            <CardDescription>Control what information is visible to others</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="flex items-center space-x-2">
              <Controller
                name="privacy_settings.show_hours"
                control={control}
                render={({ field }) => (
                  <Checkbox
                    id="show_hours"
                    checked={field.value}
                    onCheckedChange={field.onChange}
                  />
                )}
              />
              <Label htmlFor="show_hours" className="cursor-pointer">
                Show my total volunteer hours publicly
              </Label>
            </div>

            <div className="flex items-center space-x-2">
              <Controller
                name="privacy_settings.show_events"
                control={control}
                render={({ field }) => (
                  <Checkbox
                    id="show_events"
                    checked={field.value}
                    onCheckedChange={field.onChange}
                  />
                )}
              />
              <Label htmlFor="show_events" className="cursor-pointer">
                Show my attended events publicly
              </Label>
            </div>

            <div className="flex items-center space-x-2">
              <Controller
                name="privacy_settings.show_organizations"
                control={control}
                render={({ field }) => (
                  <Checkbox
                    id="show_organizations"
                    checked={field.value}
                    onCheckedChange={field.onChange}
                  />
                )}
              />
              <Label htmlFor="show_organizations" className="cursor-pointer">
                Show organizations I've volunteered with
              </Label>
            </div>
          </CardContent>
        </Card>

        {/* Notification Settings */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Bell className="h-5 w-5" />
              Notification Preferences
            </CardTitle>
            <CardDescription>Choose how you want to receive notifications</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="flex items-center space-x-2">
              <Controller
                name="notification_settings.in_app"
                control={control}
                render={({ field }) => (
                  <Checkbox
                    id="notification_in_app"
                    checked={field.value}
                    onCheckedChange={field.onChange}
                  />
                )}
              />
              <Label htmlFor="notification_in_app" className="cursor-pointer">
                In-app notifications
              </Label>
            </div>

            <div className="flex items-center space-x-2">
              <Controller
                name="notification_settings.browser_push"
                control={control}
                render={({ field }) => (
                  <Checkbox
                    id="notification_browser_push"
                    checked={field.value}
                    onCheckedChange={field.onChange}
                  />
                )}
              />
              <Label htmlFor="notification_browser_push" className="cursor-pointer">
                Browser push notifications
              </Label>
            </div>
          </CardContent>
        </Card>

        {/* Bottom Save Button */}
        <div className="flex justify-end">
          <Button type="submit" disabled={!isDirty || isSaving} className="gap-2">
            {isSaving ? (
              <>
                <Loader2 className="h-4 w-4 animate-spin" />
                Saving...
              </>
            ) : (
              <>
                <Save className="h-4 w-4" />
                Save Changes
              </>
            )}
          </Button>
        </div>
      </form>
    </div>
  );
}
