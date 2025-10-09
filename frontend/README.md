# VolunteerSync Frontend

Next.js 15 application for the VolunteerSync volunteer management platform.

## Tech Stack

- **Framework**: Next.js 15 (App Router)
- **Language**: TypeScript
- **UI**: React 19, Tailwind CSS, shadcn/ui
- **State Management**: Zustand
- **API Client**: Custom fetch wrapper with JWT token handling
- **Testing**: Jest, React Testing Library, Playwright (E2E)

## Getting Started

### Prerequisites

- Node.js 20+ or Bun
- Backend API running (see backend README)

### Environment Setup

1. **Copy the environment file**:

   ```bash
   cp .env.example .env.local
   ```

2. **Configure the backend API URL**:
   Edit `.env.local` and set:

   ```env
   NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
   ```

   **Important**: The backend runs on port **8080** (not 3000). Make sure:
   - The backend is running on `http://localhost:8080`
   - The `.env.local` file points to the correct URL
   - For Docker Compose, the backend service exposes port 8080

### Development

Run the development server:

```bash
npm run dev
# or
yarn dev
# or
pnpm dev
# or
bun dev
```

Open [http://localhost:3000](http://localhost:3000) with your browser.

### Building for Production

```bash
npm run build
npm start
```

## Project Structure

```
src/
├── app/              # Next.js App Router pages
├── components/       # React components
│   ├── ui/          # shadcn/ui components
│   ├── features/    # Feature-specific components
│   └── shared/      # Reusable components
├── lib/             # Utilities and helpers
│   ├── api/         # API client and types
│   ├── hooks/       # Custom React hooks
│   └── utils/       # Utility functions
└── store/           # Zustand stores
```

## API Configuration

The frontend communicates with the backend API at `http://localhost:8080/api/v1` by default.

### Environment Variables

| Variable              | Description          | Default                        |
| --------------------- | -------------------- | ------------------------------ |
| `NEXT_PUBLIC_API_URL` | Backend API base URL | `http://localhost:8080/api/v1` |

**Note**: `NEXT_PUBLIC_*` variables are embedded in the client-side bundle and executed in the browser. They should reference URLs accessible from the user's browser (e.g., `localhost:8080`), not internal Docker service names.

## Testing

### Unit Tests (Jest + React Testing Library)

```bash
npm test
npm run test:watch
npm run test:coverage
```

### E2E Tests (Playwright)

```bash
npm run test:e2e
npm run test:e2e:ui  # Interactive UI mode
```

## Docker Development

When running with Docker Compose:

```bash
cd docker
docker compose up frontend backend postgres redis
```

The frontend will be available at `http://localhost:3000` and will connect to the backend at `http://localhost:8080`.

## Common Issues

### "Failed to fetch" or "Network Error"

1. **Check backend is running**: `curl http://localhost:8080/health`
2. **Verify .env.local**: Ensure `NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1`
3. **Check CORS**: Backend must allow `http://localhost:3000` in CORS origins
4. **Firewall**: Ensure port 8080 is not blocked

### Port Already in Use

If port 3000 is already in use:

```bash
# Find process using port 3000
lsof -ti:3000
# Kill the process
kill -9 <PID>
```

## Learn More

- [Next.js Documentation](https://nextjs.org/docs)
- [React Documentation](https://react.dev)
- [Tailwind CSS](https://tailwindcss.com/docs)
- [shadcn/ui](https://ui.shadcn.com)
