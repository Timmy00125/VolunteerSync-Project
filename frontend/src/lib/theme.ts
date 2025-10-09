/**
 * Tailwind CSS v4 Theme Utilities
 *
 * Helper functions for accessing Tailwind theme values in JavaScript/TypeScript
 * Since Tailwind v4 uses CSS variables, we need to access them via getComputedStyle
 */

/**
 * Get a CSS variable value from the document root
 * @param variableName - The CSS variable name (with or without --)
 * @returns The computed value of the CSS variable
 */
export function getThemeVariable(variableName: string): string {
  const varName = variableName.startsWith('--') ? variableName : `--${variableName}`;
  return getComputedStyle(document.documentElement).getPropertyValue(varName).trim();
}

/**
 * Get a color value from the theme
 * @param colorName - The color name (e.g., 'primary', 'volunteer', 'organization')
 * @returns The OKLCH color value
 */
export function getThemeColor(colorName: string): string {
  return getThemeVariable(`color-${colorName}`);
}

/**
 * Get a spacing value from the theme
 * @param size - The spacing size (e.g., '18', '22', '26')
 * @returns The spacing value in rem
 */
export function getThemeSpacing(size: string | number): string {
  return getThemeVariable(`spacing-${size}`);
}

/**
 * Get a breakpoint value from the theme
 * @param breakpoint - The breakpoint name (e.g., '3xl', '4xl')
 * @returns The breakpoint value in rem
 */
export function getThemeBreakpoint(breakpoint: string): string {
  return getThemeVariable(`breakpoint-${breakpoint}`);
}

/**
 * Convert a Tailwind color variable to a CSS variable reference
 * Useful for CSS-in-JS solutions like styled-components or emotion
 * @param colorName - The color name
 * @returns CSS variable reference string
 */
export function toColorVar(colorName: string): string {
  return `var(--color-${colorName})`;
}

/**
 * Platform-specific color helpers
 */
export const platformColors = {
  volunteer: toColorVar('volunteer'),
  volunteerFg: toColorVar('volunteer-foreground'),
  organization: toColorVar('organization'),
  organizationFg: toColorVar('organization-foreground'),
  opportunity: toColorVar('opportunity'),
  opportunityFg: toColorVar('opportunity-foreground'),
  success: toColorVar('success'),
  successFg: toColorVar('success-foreground'),
  warning: toColorVar('warning'),
  warningFg: toColorVar('warning-foreground'),
  info: toColorVar('info'),
  infoFg: toColorVar('info-foreground'),
} as const;

/**
 * shadcn/ui semantic color helpers
 */
export const semanticColors = {
  background: toColorVar('background'),
  foreground: toColorVar('foreground'),
  card: toColorVar('card'),
  cardForeground: toColorVar('card-foreground'),
  popover: toColorVar('popover'),
  popoverForeground: toColorVar('popover-foreground'),
  primary: toColorVar('primary'),
  primaryForeground: toColorVar('primary-foreground'),
  secondary: toColorVar('secondary'),
  secondaryForeground: toColorVar('secondary-foreground'),
  muted: toColorVar('muted'),
  mutedForeground: toColorVar('muted-foreground'),
  accent: toColorVar('accent'),
  accentForeground: toColorVar('accent-foreground'),
  destructive: toColorVar('destructive'),
  border: toColorVar('border'),
  input: toColorVar('input'),
  ring: toColorVar('ring'),
} as const;

/**
 * Check if dark mode is currently active
 * @returns true if dark mode is active
 */
export function isDarkMode(): boolean {
  return document.documentElement.classList.contains('dark');
}

/**
 * Example usage:
 *
 * ```typescript
 * // Get a color value
 * const volunteerColor = getThemeColor('volunteer');
 *
 * // Use in inline styles
 * <div style={{ backgroundColor: platformColors.volunteer }}>
 *   Volunteer Section
 * </div>
 *
 * // Use with motion/framer-motion
 * <motion.div
 *   animate={{ backgroundColor: platformColors.organization }}
 * />
 *
 * // Get computed values
 * const spacing18 = getThemeSpacing('18'); // "4.5rem"
 * const breakpoint3xl = getThemeBreakpoint('3xl'); // "120rem"
 * ```
 */
