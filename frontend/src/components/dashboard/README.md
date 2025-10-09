# Dashboard Components

This directory contains reusable dashboard layout components for the VolunteerSync platform.

## Components

### VolunteerSidebar

**File**: `VolunteerSidebar.tsx`

Navigation sidebar for the volunteer dashboard with active route highlighting and responsive design.

**Features**:

- Icon-based navigation with labels
- Active route highlighting
- Scrollable navigation list
- Brand section with VolunteerSync logo
- Footer with motivational text

**Navigation Items**:

- Dashboard (Overview)
- Find Opportunities (Search)
- My Events (Calendar)
- Impact (Metrics)
- Profile (User settings)

**Usage**:

```tsx
import { VolunteerSidebar } from '@/components/dashboard/VolunteerSidebar';

<VolunteerSidebar />;
```

### VolunteerHeader

**File**: `VolunteerHeader.tsx`

Top header bar with notifications and user menu for the volunteer dashboard.

**Features**:

- Notifications bell with unread count badge
- User avatar with dropdown menu
- Quick access to profile and settings
- Logout functionality
- Responsive design (hides user info on mobile)

**Props**:

```tsx
interface VolunteerHeaderProps {
  unreadCount?: number; // Number of unread notifications
  user?: {
    name: string;
    email: string;
    avatar?: string;
  };
}
```

**Usage**:

```tsx
import { VolunteerHeader } from '@/components/dashboard/VolunteerHeader';

<VolunteerHeader
  unreadCount={5}
  user={{
    name: 'Jane Doe',
    email: 'jane@example.com',
    avatar: '/avatars/jane.jpg',
  }}
/>;
```

## Layout Structure

The volunteer dashboard uses a fixed layout structure:

```
┌─────────────────────────────────────────┐
│             VolunteerHeader             │
│  (Notifications, User Menu)             │
├─────────┬───────────────────────────────┤
│         │                               │
│ Sidebar │      Main Content Area        │
│ (Nav)   │      (Page Children)          │
│         │                               │
│         │                               │
└─────────┴───────────────────────────────┘
```

**Dimensions**:

- Sidebar: 256px (16rem) width, full height
- Header: 64px (4rem) height, full width
- Content: Fills remaining space with scroll

## Dependencies

The dashboard components use the following shadcn/ui components:

- `avatar` - User profile picture
- `badge` - Notification count indicator
- `button` - Interactive elements
- `dropdown-menu` - User and notification menus
- `scroll-area` - Scrollable navigation list

## Styling

Components follow the design system defined in:

- `frontend/src/app/globals.css` - Global styles
- `frontend/components.json` - shadcn/ui configuration (New York style)
- Tailwind CSS utility classes with zinc base color

## Future Enhancements

### Mobile Menu

- Add hamburger menu toggle for mobile devices
- Implement slide-in overlay sidebar for small screens
- Add touch gestures for mobile navigation

### User Context Integration

- Connect to authentication context for real user data
- Fetch unread notification count from API
- Implement actual logout flow with token invalidation

### Accessibility

- Add keyboard navigation shortcuts
- Improve screen reader announcements
- Add focus management for dropdown menus

## Related Files

- Layout: `frontend/src/app/(dashboard)/volunteer/layout.tsx`
- Dashboard page: `frontend/src/app/(dashboard)/volunteer/page.tsx`
- Task reference: Task T107 in `specs/001-build-volunteersync-an/tasks.md`
