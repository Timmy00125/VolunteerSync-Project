/**
 * Example Components Using Tailwind CSS v4 Custom Theme
 *
 * This file demonstrates how to use the custom theme configuration
 * in real React components for the VolunteerSync platform.
 */

import React from 'react';
import { platformColors } from '@/lib/theme';

/**
 * Example 1: Volunteer Dashboard Card
 * Uses custom volunteer colors and shadow utilities
 */
export function VolunteerCard() {
  return (
    <div className="bg-card shadow-card hover:shadow-card-hover transition-shadow rounded-xl p-6">
      <div className="flex items-center gap-4 mb-4">
        <div className="w-12 h-12 rounded-full bg-volunteer text-volunteer-foreground flex items-center justify-center font-bold text-xl">
          JD
        </div>
        <div>
          <h3 className="font-display text-lg font-semibold">John Doe</h3>
          <p className="text-muted-foreground text-sm">Active Volunteer</p>
        </div>
      </div>

      <div className="space-y-2">
        <div className="flex justify-between items-center">
          <span className="text-sm text-muted-foreground">Total Hours</span>
          <span className="font-semibold text-volunteer">127 hrs</span>
        </div>
        <div className="flex justify-between items-center">
          <span className="text-sm text-muted-foreground">Events Completed</span>
          <span className="font-semibold text-success">23</span>
        </div>
      </div>
    </div>
  );
}

/**
 * Example 2: Organization Banner
 * Uses custom organization colors and custom breakpoints
 */
export function OrganizationBanner() {
  return (
    <div className="bg-organization text-organization-foreground p-6 rounded-2xl">
      <div className="grid grid-cols-1 md:grid-cols-2 3xl:grid-cols-4 gap-6">
        <div>
          <h4 className="text-2xs uppercase tracking-wide opacity-80 mb-1">Active Opportunities</h4>
          <p className="text-3xl font-display font-bold">12</p>
        </div>
        <div>
          <h4 className="text-2xs uppercase tracking-wide opacity-80 mb-1">Total Volunteers</h4>
          <p className="text-3xl font-display font-bold">48</p>
        </div>
        <div>
          <h4 className="text-2xs uppercase tracking-wide opacity-80 mb-1">Hours This Month</h4>
          <p className="text-3xl font-display font-bold">340</p>
        </div>
        <div>
          <h4 className="text-2xs uppercase tracking-wide opacity-80 mb-1">Impact Score</h4>
          <p className="text-3xl font-display font-bold">92%</p>
        </div>
      </div>
    </div>
  );
}

/**
 * Example 3: Opportunity Card
 * Uses custom opportunity colors and custom animations
 */
export function OpportunityCard() {
  const [isNew, setIsNew] = React.useState(true);

  React.useEffect(() => {
    const timer = setTimeout(() => setIsNew(false), 3000);
    return () => clearTimeout(timer);
  }, []);

  return (
    <div className="bg-card border border-border rounded-xl overflow-hidden hover:border-opportunity transition-colors">
      {isNew && (
        <div className="bg-opportunity text-opportunity-foreground px-4 py-2 text-sm font-medium animate-fade-in">
          ✨ New Opportunity
        </div>
      )}

      <div className="p-6">
        <h3 className="font-display text-xl font-semibold mb-2">Community Garden Planting</h3>
        <p className="text-muted-foreground mb-4">
          Join us for a day of planting flowers and vegetables in our community garden.
        </p>

        <div className="flex gap-2 mb-4">
          <span className="bg-success/10 text-success px-3 py-1 rounded-lg text-sm">
            Environment
          </span>
          <span className="bg-info/10 text-info px-3 py-1 rounded-lg text-sm">Outdoor</span>
        </div>

        <button className="w-full bg-opportunity text-opportunity-foreground py-2 rounded-lg font-medium hover:opacity-90 transition-opacity">
          Register Now
        </button>
      </div>
    </div>
  );
}

/**
 * Example 4: Alert Messages
 * Uses custom success, warning, and info colors
 */
export function AlertExamples() {
  return (
    <div className="space-y-4">
      {/* Success Alert */}
      <div className="bg-success/10 border border-success text-success p-4 rounded-xl animate-slide-in-right">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-full bg-success text-success-foreground flex items-center justify-center">
            ✓
          </div>
          <div>
            <h4 className="font-semibold">Successfully Registered!</h4>
            <p className="text-sm opacity-80">You're all set for the event on Saturday.</p>
          </div>
        </div>
      </div>

      {/* Warning Alert */}
      <div className="bg-warning/10 border border-warning text-warning p-4 rounded-xl">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-full bg-warning text-warning-foreground flex items-center justify-center">
            ⚠
          </div>
          <div>
            <h4 className="font-semibold">Late Cancellation</h4>
            <p className="text-sm opacity-80">
              Canceling within 24 hours may affect your reliability score.
            </p>
          </div>
        </div>
      </div>

      {/* Info Alert */}
      <div className="bg-info/10 border border-info text-info p-4 rounded-xl">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-full bg-info text-info-foreground flex items-center justify-center">
            ℹ
          </div>
          <div>
            <h4 className="font-semibold">Pro Tip</h4>
            <p className="text-sm opacity-80">
              Complete your profile to get better opportunity matches.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}

/**
 * Example 5: Dark Mode Toggle
 * Demonstrates dark mode support
 */
export function DarkModeToggle() {
  const [isDark, setIsDark] = React.useState(false);

  const toggleDarkMode = () => {
    setIsDark(!isDark);
    document.documentElement.classList.toggle('dark');
  };

  return (
    <button
      onClick={toggleDarkMode}
      className="p-3 rounded-xl bg-muted hover:bg-muted/80 transition-colors"
      aria-label="Toggle dark mode"
    >
      {isDark ? '🌙' : '☀️'}
    </button>
  );
}

/**
 * Example 6: Stat Card with Custom Spacing
 * Uses custom spacing utilities (spacing-18, spacing-22, etc.)
 */
export function StatCard() {
  return (
    <div className="bg-gradient-to-br from-primary to-primary/80 text-primary-foreground p-6 rounded-3xl shadow-3xl">
      <div className="space-y-2">
        <p className="text-2xs uppercase tracking-wide opacity-80">Your Impact</p>
        <p className="text-5xl font-display font-bold">127</p>
        <p className="text-sm opacity-90">volunteer hours this year</p>
      </div>

      <div className="mt-18 pt-4 border-t border-white/20">
        <p className="text-sm">
          You've made a difference in <span className="font-bold">12 communities</span>
        </p>
      </div>
    </div>
  );
}

/**
 * Example 7: With Motion Animation
 * Using theme colors with framer-motion
 */
export function AnimatedCard() {
  // Note: This requires framer-motion to be installed
  // import { motion } from 'framer-motion';

  return (
    <div
      // motion.div
      // initial={{ opacity: 0, y: 20 }}
      // animate={{
      //   opacity: 1,
      //   y: 0,
      //   backgroundColor: platformColors.volunteer
      // }}
      // transition={{ duration: 0.3 }}
      className="p-6 rounded-xl text-volunteer-foreground"
      style={{ backgroundColor: platformColors.volunteer }}
    >
      <h3 className="font-display text-xl font-semibold">Animated with Theme Colors</h3>
      <p className="text-sm opacity-90 mt-2">
        This card animates using CSS variables from the theme!
      </p>
    </div>
  );
}

/**
 * Example 8: Responsive Grid with Custom Breakpoints
 * Shows 3xl and 4xl breakpoints in action
 */
export function ResponsiveGrid() {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 3xl:grid-cols-4 4xl:grid-cols-6 gap-6">
      {Array.from({ length: 12 }).map((_, i) => (
        <div
          key={i}
          className="bg-card border border-border p-6 rounded-xl aspect-square flex items-center justify-center"
        >
          <span className="text-muted-foreground">Item {i + 1}</span>
        </div>
      ))}
    </div>
  );
}

/**
 * Example 9: Badge Components
 * Using various theme colors for badges
 */
export function BadgeExamples() {
  return (
    <div className="flex flex-wrap gap-2">
      <span className="bg-volunteer text-volunteer-foreground px-3 py-1 rounded-full text-sm font-medium">
        Volunteer
      </span>
      <span className="bg-organization text-organization-foreground px-3 py-1 rounded-full text-sm font-medium">
        Organization
      </span>
      <span className="bg-opportunity text-opportunity-foreground px-3 py-1 rounded-full text-sm font-medium">
        Opportunity
      </span>
      <span className="bg-success text-success-foreground px-3 py-1 rounded-full text-sm font-medium">
        Verified
      </span>
      <span className="bg-warning text-warning-foreground px-3 py-1 rounded-full text-sm font-medium">
        Pending
      </span>
      <span className="bg-info text-info-foreground px-3 py-1 rounded-full text-sm font-medium">
        New
      </span>
    </div>
  );
}

/**
 * Example 10: Complete Dashboard Layout
 * Combining multiple custom theme features
 */
export function DashboardExample() {
  return (
    <div className="min-h-screen bg-background p-6">
      <div className="max-w-7xl mx-auto space-y-6">
        {/* Header */}
        <div className="flex justify-between items-center mb-22">
          <h1 className="font-display text-4xl font-bold text-foreground">Dashboard</h1>
          <DarkModeToggle />
        </div>

        {/* Stats Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 3xl:grid-cols-4 gap-6 mb-18">
          <StatCard />
          {/* Add more stat cards */}
        </div>

        {/* Alerts */}
        <AlertExamples />

        {/* Content Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mt-18">
          <div className="lg:col-span-2">
            <h2 className="font-display text-2xl font-semibold mb-4">Upcoming Opportunities</h2>
            <div className="space-y-4">
              <OpportunityCard />
              <OpportunityCard />
            </div>
          </div>

          <div>
            <h2 className="font-display text-2xl font-semibold mb-4">Your Profile</h2>
            <VolunteerCard />
          </div>
        </div>
      </div>
    </div>
  );
}

/**
 * To use these examples:
 *
 * 1. Copy the component you need
 * 2. Import it in your page/component
 * 3. Customize as needed
 *
 * All components use the custom theme configured in globals.css
 * and will automatically adapt to light/dark mode.
 */
