/**
 * HoursChart Component
 *
 * Simple bar chart visualization for displaying volunteer hours over time.
 * This is a basic implementation using SVG. Can be enhanced with Chart.js or Recharts later.
 */

'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

interface HoursChartData {
  label: string;
  hours: number;
}

interface HoursChartProps {
  /**
   * Data points for the chart
   */
  data: HoursChartData[];

  /**
   * Chart title
   */
  title?: string;

  /**
   * Optional className for custom styling
   */
  className?: string;
}

/**
 * HoursChart - Display volunteer hours over time as a bar chart
 *
 * @example
 * ```tsx
 * <HoursChart
 *   title="Hours This Year"
 *   data={[
 *     { label: 'Jan', hours: 5 },
 *     { label: 'Feb', hours: 8 },
 *     { label: 'Mar', hours: 12 },
 *   ]}
 * />
 * ```
 */
export function HoursChart({ data, title = 'Hours Over Time', className }: HoursChartProps) {
  if (!data || data.length === 0) {
    return (
      <Card className={className}>
        <CardHeader>
          <CardTitle>{title}</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">No data available</p>
        </CardContent>
      </Card>
    );
  }

  // Find max hours for scaling
  const maxHours = Math.max(...data.map((d) => d.hours), 1);
  const chartHeight = 200;
  const barWidth = 100 / data.length;

  return (
    <Card className={className}>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="relative" style={{ height: `${chartHeight + 40}px` }}>
          {/* Chart area */}
          <div
            className="relative flex items-end justify-around gap-2"
            style={{ height: `${chartHeight}px` }}
          >
            {data.map((item, index) => {
              const barHeight = (item.hours / maxHours) * chartHeight;
              return (
                <div
                  key={index}
                  className="flex flex-col items-center gap-1"
                  style={{ width: `${barWidth}%` }}
                >
                  {/* Bar */}
                  <div
                    className="relative flex w-full flex-col items-center justify-end"
                    style={{ height: `${chartHeight}px` }}
                  >
                    {/* Hours label on top of bar */}
                    <div className="mb-1 text-xs font-medium text-primary">
                      {item.hours > 0 ? item.hours.toFixed(1) : ''}
                    </div>
                    {/* Bar itself */}
                    <div
                      className="w-full rounded-t-sm bg-primary transition-all hover:opacity-80"
                      style={{
                        height: `${barHeight}px`,
                        minHeight: item.hours > 0 ? '4px' : '0px',
                      }}
                      title={`${item.label}: ${item.hours} hours`}
                    />
                  </div>
                </div>
              );
            })}
          </div>

          {/* X-axis labels */}
          <div className="mt-2 flex items-center justify-around">
            {data.map((item, index) => (
              <div
                key={index}
                className="text-center text-xs text-muted-foreground"
                style={{ width: `${barWidth}%` }}
              >
                {item.label}
              </div>
            ))}
          </div>
        </div>

        {/* Summary */}
        <div className="mt-4 flex items-center justify-between border-t pt-4">
          <span className="text-sm text-muted-foreground">Total</span>
          <span className="text-sm font-medium">
            {data.reduce((sum, item) => sum + item.hours, 0).toFixed(1)} hours
          </span>
        </div>
      </CardContent>
    </Card>
  );
}
