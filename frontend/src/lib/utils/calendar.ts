/**
 * Calendar Utilities
 *
 * Provides functions to generate calendar files and links for events.
 * Supports:
 * - .ics file generation for calendar applications
 * - Google Calendar links
 * - Download functionality
 */

export interface CalendarEvent {
  title: string;
  description: string;
  location: string;
  startTime: Date | string;
  endTime: Date | string;
  url?: string;
  organizer?: {
    name: string;
    email: string;
  };
  attendee?: {
    name: string;
    email: string;
  };
}

/**
 * Format a date to ICS format (YYYYMMDDTHHmmssZ)
 * @param date - Date object or ISO string
 * @returns Formatted date string for ICS
 */
function formatICSDate(date: Date | string): string {
  const d = typeof date === 'string' ? new Date(date) : date;

  // Format: YYYYMMDDTHHmmssZ
  const year = d.getUTCFullYear();
  const month = String(d.getUTCMonth() + 1).padStart(2, '0');
  const day = String(d.getUTCDate()).padStart(2, '0');
  const hours = String(d.getUTCHours()).padStart(2, '0');
  const minutes = String(d.getUTCMinutes()).padStart(2, '0');
  const seconds = String(d.getUTCSeconds()).padStart(2, '0');

  return `${year}${month}${day}T${hours}${minutes}${seconds}Z`;
}

/**
 * Escape special characters for ICS format
 * @param text - Text to escape
 * @returns Escaped text
 */
function escapeICSText(text: string): string {
  return text
    .replace(/\\/g, '\\\\')
    .replace(/;/g, '\\;')
    .replace(/,/g, '\\,')
    .replace(/\n/g, '\\n');
}

/**
 * Generate a unique identifier for the event
 * @returns Unique ID
 */
function generateUID(): string {
  return `${Date.now()}-${Math.random().toString(36).substring(2, 9)}@volunteersync.com`;
}

/**
 * Generate an .ics file content for a calendar event
 * @param event - Event details
 * @returns ICS file content as string
 */
export function generateICS(event: CalendarEvent): string {
  const startDate = formatICSDate(event.startTime);
  const endDate = formatICSDate(event.endTime);
  const now = formatICSDate(new Date());
  const uid = generateUID();

  let icsContent = [
    'BEGIN:VCALENDAR',
    'VERSION:2.0',
    'PRODID:-//VolunteerSync//Calendar//EN',
    'CALSCALE:GREGORIAN',
    'METHOD:PUBLISH',
    'BEGIN:VEVENT',
    `UID:${uid}`,
    `DTSTAMP:${now}`,
    `DTSTART:${startDate}`,
    `DTEND:${endDate}`,
    `SUMMARY:${escapeICSText(event.title)}`,
    `DESCRIPTION:${escapeICSText(event.description)}`,
    `LOCATION:${escapeICSText(event.location)}`,
  ];

  // Add URL if provided
  if (event.url) {
    icsContent.push(`URL:${event.url}`);
  }

  // Add organizer if provided
  if (event.organizer) {
    icsContent.push(
      `ORGANIZER;CN=${escapeICSText(event.organizer.name)}:MAILTO:${event.organizer.email}`
    );
  }

  // Add attendee if provided
  if (event.attendee) {
    icsContent.push(
      `ATTENDEE;CN=${escapeICSText(event.attendee.name)};ROLE=REQ-PARTICIPANT;PARTSTAT=NEEDS-ACTION;RSVP=TRUE:MAILTO:${event.attendee.email}`
    );
  }

  icsContent.push('STATUS:CONFIRMED', 'SEQUENCE:0', 'END:VEVENT', 'END:VCALENDAR');

  return icsContent.join('\r\n');
}

/**
 * Download an .ics file
 * @param event - Event details
 * @param filename - Optional custom filename (default: event title)
 */
export function downloadICS(event: CalendarEvent, filename?: string): void {
  const icsContent = generateICS(event);
  const blob = new Blob([icsContent], { type: 'text/calendar;charset=utf-8' });
  const url = URL.createObjectURL(blob);

  const link = document.createElement('a');
  link.href = url;
  link.download = filename || `${event.title.replace(/[^a-z0-9]/gi, '_').toLowerCase()}.ics`;

  // Trigger download
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);

  // Clean up
  URL.revokeObjectURL(url);
}

/**
 * Generate a Google Calendar link
 * @param event - Event details
 * @returns Google Calendar URL
 */
export function generateGoogleCalendarLink(event: CalendarEvent): string {
  const startDate =
    typeof event.startTime === 'string' ? new Date(event.startTime) : event.startTime;
  const endDate = typeof event.endTime === 'string' ? new Date(event.endTime) : event.endTime;

  // Format dates for Google Calendar (YYYYMMDDTHHmmssZ)
  const formatGoogleDate = (date: Date) => {
    return formatICSDate(date);
  };

  const params = new URLSearchParams({
    action: 'TEMPLATE',
    text: event.title,
    details: event.description,
    location: event.location,
    dates: `${formatGoogleDate(startDate)}/${formatGoogleDate(endDate)}`,
  });

  // Add URL to details if provided
  if (event.url) {
    params.set('details', `${event.description}\n\nMore info: ${event.url}`);
  }

  return `https://calendar.google.com/calendar/render?${params.toString()}`;
}

/**
 * Open Google Calendar in a new tab/window
 * @param event - Event details
 */
export function openGoogleCalendar(event: CalendarEvent): void {
  const url = generateGoogleCalendarLink(event);
  window.open(url, '_blank', 'noopener,noreferrer');
}

/**
 * Generate an Outlook Calendar link
 * @param event - Event details
 * @returns Outlook Calendar URL
 */
export function generateOutlookCalendarLink(event: CalendarEvent): string {
  const startDate =
    typeof event.startTime === 'string' ? new Date(event.startTime) : event.startTime;
  const endDate = typeof event.endTime === 'string' ? new Date(event.endTime) : event.endTime;

  const params = new URLSearchParams({
    path: '/calendar/action/compose',
    rru: 'addevent',
    subject: event.title,
    body: event.description,
    location: event.location,
    startdt: startDate.toISOString(),
    enddt: endDate.toISOString(),
  });

  return `https://outlook.live.com/calendar/0/deeplink/compose?${params.toString()}`;
}

/**
 * Open Outlook Calendar in a new tab/window
 * @param event - Event details
 */
export function openOutlookCalendar(event: CalendarEvent): void {
  const url = generateOutlookCalendarLink(event);
  window.open(url, '_blank', 'noopener,noreferrer');
}

/**
 * Generate an Apple Calendar (iCal) link
 * This generates a data URL that can be used as a download link
 * @param event - Event details
 * @returns Data URL for .ics file
 */
export function generateAppleCalendarLink(event: CalendarEvent): string {
  const icsContent = generateICS(event);
  const blob = new Blob([icsContent], { type: 'text/calendar;charset=utf-8' });
  return URL.createObjectURL(blob);
}

/**
 * Detect user's calendar preference (if possible)
 * @returns Suggested calendar type
 */
export function detectCalendarPreference(): 'google' | 'outlook' | 'apple' | 'ics' {
  if (typeof window === 'undefined') {
    return 'ics';
  }

  const userAgent = window.navigator.userAgent.toLowerCase();

  // Check for macOS/iOS
  if (userAgent.includes('mac') || userAgent.includes('iphone') || userAgent.includes('ipad')) {
    return 'apple';
  }

  // Check for Windows
  if (userAgent.includes('windows')) {
    return 'outlook';
  }

  // Default to Google Calendar for others
  return 'google';
}

/**
 * Add event to calendar based on user's detected preference
 * @param event - Event details
 */
export function addToCalendar(event: CalendarEvent): void {
  const preference = detectCalendarPreference();

  switch (preference) {
    case 'google':
      openGoogleCalendar(event);
      break;
    case 'outlook':
      openOutlookCalendar(event);
      break;
    case 'apple':
    case 'ics':
    default:
      downloadICS(event);
      break;
  }
}
