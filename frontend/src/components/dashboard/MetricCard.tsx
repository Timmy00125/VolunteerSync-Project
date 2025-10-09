/**
 * MetricCard Component
 *
 * Reusable card component for displaying dashboard metrics with an icon.
 * Used to show statistics like total hours, events, organizations, etc.
 */

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { LucideIcon } from 'lucide-react';

interface MetricCardProps {
  /**
   * Title/label for the metric
   */
  title: string;

  /**
   * Main metric value to display
   */
  value: string | number;

  /**
   * Icon component from lucide-react
   */
  icon: LucideIcon;

  /**
   * Optional subtitle or additional info
   */
  subtitle?: string;

  /**
   * Optional className for custom styling
   */
  className?: string;
}

/**
 * MetricCard - Display a metric with an icon
 *
 * @example
 * ```tsx
 * <MetricCard
 *   title="Total Hours"
 *   value={24.5}
 *   icon={Clock}
 *   subtitle="+4.5 hours this month"
 * />
 * ```
 */
export function MetricCard({ title, value, icon: Icon, subtitle, className }: MetricCardProps) {
  return (
    <Card className={className}>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">{title}</CardTitle>
        <Icon className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">{value}</div>
        {subtitle && <p className="text-xs text-muted-foreground">{subtitle}</p>}
      </CardContent>
    </Card>
  );
}
