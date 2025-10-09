'use client';

import { useState, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useOrganizationTeam, useInviteTeamMember, useRemoveTeamMember } from '@/lib/api';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Users,
  UserPlus,
  Mail,
  Trash2,
  Shield,
  Briefcase,
  CheckCircle,
  AlertCircle,
} from 'lucide-react';
import { format, parseISO } from 'date-fns';
import type { TeamMember } from '@/lib/api/types';

/**
 * Team Management Page (T121)
 *
 * Manage organization team members:
 * - List team members with roles
 * - Invite new members (send email invitation)
 * - Remove members
 *
 * Roles:
 * - admin: Full access to organization settings
 * - coordinator: Can create opportunities and manage volunteers
 * - volunteer: Basic member (not shown on this page)
 */

const inviteSchema = z.object({
  email: z.string().email('Must be a valid email address'),
  role: z.enum(['admin', 'coordinator'], {
    message: 'Please select a role',
  }),
});

type InviteFormData = z.infer<typeof inviteSchema>;

export default function TeamManagementPage() {
  // TODO: Get organization ID from auth context
  const [organizationId, setOrganizationId] = useState<string>('');
  const [showInviteForm, setShowInviteForm] = useState(false);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  useEffect(() => {
    // Placeholder: would get from useAuth() hook or similar
    setOrganizationId('org-1');
  }, []);

  const { data: teamMembers, isLoading } = useOrganizationTeam(organizationId);
  const inviteMutation = useInviteTeamMember(organizationId);
  const removeMutation = useRemoveTeamMember(organizationId);

  const {
    register,
    handleSubmit,
    setValue,
    watch,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<InviteFormData>({
    resolver: zodResolver(inviteSchema),
    defaultValues: {
      role: 'coordinator',
    },
  });

  const selectedRole = watch('role');

  const onInvite = async (data: InviteFormData) => {
    try {
      setErrorMessage(null);
      await inviteMutation.mutateAsync(data);
      setSuccessMessage(`Invitation sent to ${data.email}`);
      setShowInviteForm(false);
      reset();
      setTimeout(() => setSuccessMessage(null), 5000);
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : 'Failed to send invitation');
    }
  };

  const handleRemoveMember = async (userId: string, memberName: string) => {
    if (!confirm(`Are you sure you want to remove ${memberName} from the team?`)) {
      return;
    }

    try {
      setErrorMessage(null);
      await removeMutation.mutateAsync(userId);
      setSuccessMessage(`${memberName} has been removed from the team`);
      setTimeout(() => setSuccessMessage(null), 5000);
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : 'Failed to remove team member');
    }
  };

  // Group members by role
  const admins = teamMembers?.filter((m) => m.role === 'admin') || [];
  const coordinators = teamMembers?.filter((m) => m.role === 'coordinator') || [];

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Team Management</h1>
          <p className="text-muted-foreground">
            Manage your organization&apos;s team members and roles
          </p>
        </div>
        <Button onClick={() => setShowInviteForm(!showInviteForm)}>
          <UserPlus className="mr-2 h-4 w-4" />
          Invite Member
        </Button>
      </div>

      {/* Success Message */}
      {successMessage && (
        <Card className="border-green-500">
          <CardContent className="flex items-center gap-2 py-4">
            <CheckCircle className="h-5 w-5 text-green-600" />
            <p className="text-sm text-green-600">{successMessage}</p>
          </CardContent>
        </Card>
      )}

      {/* Error Message */}
      {errorMessage && (
        <Card className="border-destructive">
          <CardContent className="flex items-center gap-2 py-4">
            <AlertCircle className="h-5 w-5 text-destructive" />
            <p className="text-sm text-destructive">{errorMessage}</p>
          </CardContent>
        </Card>
      )}

      {/* Invite Form */}
      {showInviteForm && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Mail className="h-5 w-5" />
              Invite Team Member
            </CardTitle>
            <CardDescription>Send an email invitation to join your organization</CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit(onInvite)} className="space-y-4">
              <div className="grid gap-4 md:grid-cols-2">
                {/* Email */}
                <div className="space-y-2">
                  <Label htmlFor="email">
                    Email Address <span className="text-destructive">*</span>
                  </Label>
                  <Input
                    id="email"
                    type="email"
                    placeholder="member@example.com"
                    {...register('email')}
                    aria-invalid={!!errors.email}
                  />
                  {errors.email && (
                    <p className="text-sm text-destructive">{errors.email.message}</p>
                  )}
                </div>

                {/* Role */}
                <div className="space-y-2">
                  <Label htmlFor="role">
                    Role <span className="text-destructive">*</span>
                  </Label>
                  <Select
                    value={selectedRole}
                    onValueChange={(value) => setValue('role', value as any)}
                  >
                    <SelectTrigger id="role">
                      <SelectValue placeholder="Select role" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="admin">
                        <div className="flex items-center gap-2">
                          <Shield className="h-4 w-4" />
                          <span>Admin - Full access</span>
                        </div>
                      </SelectItem>
                      <SelectItem value="coordinator">
                        <div className="flex items-center gap-2">
                          <Briefcase className="h-4 w-4" />
                          <span>Coordinator - Manage opportunities</span>
                        </div>
                      </SelectItem>
                    </SelectContent>
                  </Select>
                  {errors.role && <p className="text-sm text-destructive">{errors.role.message}</p>}
                </div>
              </div>

              {/* Role Description */}
              <div className="rounded-lg bg-muted p-3 text-sm text-muted-foreground">
                {selectedRole === 'admin' ? (
                  <p>
                    <strong>Admin</strong> - Can manage organization settings, team members, and all
                    opportunities
                  </p>
                ) : (
                  <p>
                    <strong>Coordinator</strong> - Can create opportunities, manage volunteers, and
                    log hours
                  </p>
                )}
              </div>

              <div className="flex items-center gap-2">
                <Button type="submit" disabled={isSubmitting}>
                  {isSubmitting ? 'Sending...' : 'Send Invitation'}
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => {
                    setShowInviteForm(false);
                    reset();
                  }}
                >
                  Cancel
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>
      )}

      {/* Team Stats */}
      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Total Members
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{teamMembers?.length || 0}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Admins</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{admins.length}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Coordinators
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{coordinators.length}</div>
          </CardContent>
        </Card>
      </div>

      {/* Admins */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Shield className="h-5 w-5" />
            Administrators ({admins.length})
          </CardTitle>
          <CardDescription>Team members with full access to organization settings</CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <p className="text-center text-muted-foreground">Loading...</p>
          ) : admins.length === 0 ? (
            <p className="text-center text-sm text-muted-foreground">No administrators</p>
          ) : (
            <div className="space-y-3">
              {admins.map((member) => (
                <MemberCard
                  key={member.id}
                  member={member}
                  onRemove={() =>
                    handleRemoveMember(member.user_id, `${member.first_name} ${member.last_name}`)
                  }
                />
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Coordinators */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Briefcase className="h-5 w-5" />
            Coordinators ({coordinators.length})
          </CardTitle>
          <CardDescription>Team members who can manage volunteer opportunities</CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <p className="text-center text-muted-foreground">Loading...</p>
          ) : coordinators.length === 0 ? (
            <p className="text-center text-sm text-muted-foreground">No coordinators</p>
          ) : (
            <div className="space-y-3">
              {coordinators.map((member) => (
                <MemberCard
                  key={member.id}
                  member={member}
                  onRemove={() =>
                    handleRemoveMember(member.user_id, `${member.first_name} ${member.last_name}`)
                  }
                />
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

/**
 * Member Card Component
 */
interface MemberCardProps {
  member: TeamMember;
  onRemove: () => void;
}

function MemberCard({ member, onRemove }: MemberCardProps) {
  return (
    <div className="flex items-center justify-between rounded-lg border p-4">
      <div className="flex items-center gap-4">
        <div className="flex h-12 w-12 items-center justify-center rounded-full bg-muted">
          <Users className="h-6 w-6 text-muted-foreground" />
        </div>
        <div>
          <p className="font-medium">
            {member.first_name} {member.last_name}
          </p>
          <p className="text-sm text-muted-foreground">{member.email}</p>
          <p className="text-xs text-muted-foreground">
            Joined: {format(parseISO(member.joined_at), 'MMM d, yyyy')}
          </p>
        </div>
      </div>
      <div className="flex items-center gap-2">
        <span
          className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${
            member.role === 'admin' ? 'bg-purple-50 text-purple-700' : 'bg-blue-50 text-blue-700'
          }`}
        >
          {member.role === 'admin' ? (
            <>
              <Shield className="mr-1 inline h-3 w-3" />
              Admin
            </>
          ) : (
            <>
              <Briefcase className="mr-1 inline h-3 w-3" />
              Coordinator
            </>
          )}
        </span>
        <Button size="sm" variant="ghost" onClick={onRemove}>
          <Trash2 className="h-4 w-4 text-destructive" />
        </Button>
      </div>
    </div>
  );
}
