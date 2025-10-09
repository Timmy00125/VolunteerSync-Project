# Backend Service Architecture

## Current Service Dependency Graph

```
┌─────────────────────────────────────────────────────────────────┐
│                         HTTP Layer (Gin)                         │
│                     Handlers + Middleware                        │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                        Service Layer                             │
│                   (Business Logic)                               │
│                                                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │ Auth Service │  │ User Service │  │ Org Service  │         │
│  │   (READY)    │  │   (READY)    │  │   (READY)    │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
│                                                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │ Vol Service  │  │ Opp Service  │  │ Reg Service  │         │
│  │   (READY)    │  │   (READY)    │  │ (NEEDS ADAPT)│         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
│                                                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │ Hours Service│  │ Comm Service │  │Analytics Svc │         │
│  │(NEEDS ADAPT) │  │(NEEDS ADAPT) │  │   (READY)    │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
│                                                                  │
│  ┌──────────────┐                                               │
│  │Achievement   │                                               │
│  │   Service    │                                               │
│  │   (READY)    │                                               │
│  └──────────────┘                                               │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Repository Layer                            │
│                    (Data Access - GORM)                          │
│                                                                  │
│  AuthRepo  │  OrgRepo  │  VolunteerRepo  │  OppRepo            │
│  RegRepo   │  HoursRepo│  CommRepo       │  AchievementRepo    │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Database Layer                                │
│                PostgreSQL 16 + Redis                             │
└─────────────────────────────────────────────────────────────────┘
```

## Service Initialization Order

```
1. Infrastructure
   ├── Logger (appLogger)
   ├── Database (PostgreSQL via GORM)
   ├── Cache (Redis)
   └── JWT Manager

2. Repositories (all require *gorm.DB)
   ├── authRepo
   ├── orgRepo
   ├── volunteerRepo
   ├── oppRepo
   ├── regRepo
   ├── hoursRepo
   ├── commRepo
   └── achievementRepo

3. Services (dependency order)
   ├── authService       (authRepo, jwtManager, redis) ✅
   ├── userService       (authRepo, DB) ✅
   ├── orgService        (orgRepo, nil, logger) ✅
   ├── volunteerService  (volunteerRepo, nil, logger) ✅
   ├── oppService        (oppRepo, nil, nil, logger) ✅
   ├── regService        (regRepo, nil, nil, logger) ⚠️
   ├── hoursService      (hoursRepo, nil, nil, nil, logger) ⚠️
   ├── commService       (commRepo, nil, logger) ⚠️
   ├── analyticsService  (DB, logger) ✅
   └── achievementService(achievementRepo, logger) ✅

4. Handlers (each requires their service)
   ├── authHandler
   ├── userHandler
   ├── orgHandler
   ├── volunteerHandler
   ├── oppHandler
   ├── regHandler
   ├── hoursHandler
   ├── commHandler
   ├── analyticsHandler
   └── achievementHandler

5. Routes (registered with Gin router)
   └── All handlers registered to appropriate route groups
```

## Module Status

### ✅ Fully Operational (No dependencies on other services)

- **Auth Module**: Login, register, logout, refresh tokens, password reset
- **Users Module**: Get/update user profile, delete account
- **Organizations Module**: Create, get, list, update organizations
- **Volunteers Module**: Get/update volunteer profile, dashboard, analytics
- **Analytics Module**: Volunteer, organization, and platform analytics
- **Achievements Module**: Badge system, award achievements

### ⚠️ Partially Operational (Need cross-module adapters)

- **Opportunities Module**: CRUD operations work, but:

  - Missing notification service for cancellations
  - Missing geocoding service for address → lat/lng

- **Registrations Module**: Basic operations work, but:

  - Missing opportunity service adapter for capacity checks
  - Missing notification service for confirmations

- **Hours Module**: Basic operations work, but:

  - Missing registration service adapter for validation
  - Missing volunteer service adapter for hour increments
  - Missing notification service for hour notifications

- **Communications Module**: Basic operations work, but:
  - Missing registration repository method for broadcast messages

## Cross-Module Dependencies (To Be Implemented)

### Priority 1: Registration ↔ Opportunity

**Why**: Core user flow (volunteer registration)

```go
// registrations/services needs:
type OpportunityService interface {
    GetOpportunity(ctx, id) (*OpportunityDetails, error)
}

// Implementation: Create adapter that calls oppService
```

### Priority 2: Hours ↔ Registration & Volunteer

**Why**: Hour tracking workflow

```go
// hours/services needs:
type RegistrationService interface {
    GetRegistration(ctx, id) (*RegistrationDetails, error)
    UpdateRegistrationHours(ctx, id, hours, status) error
}

type VolunteerService interface {
    IncrementTotalHours(ctx, volunteerProfileID, hours) error
}

// Implementation: Create adapters that call regService and volService
```

### Priority 3: Communications ↔ Registration

**Why**: Broadcast messages to volunteers

```go
// communications/services needs:
type RegistrationRepository interface {
    FindVolunteersByOpportunity(ctx, oppID) ([]uuid.UUID, error)
}

// Implementation: Add method to registrations/repositories
```

### Optional: Notification Service (All Modules)

**Why**: User notifications across the platform

- Can be implemented as a separate module
- Or integrated into communications module
- Multiple modules need to send notifications

### Optional: Geocoding Service (Org, Vol, Opp)

**Why**: Address → lat/lng conversion

- Can use external API (Google Maps, Mapbox, etc.)
- Or implement simple internal geocoding
- Currently passed as nil (optional)

## Testing Strategy

### 1. Unit Tests (Per Service)

Each service should have unit tests with mocked dependencies:

```bash
go test ./internal/modules/*/services/...
```

### 2. Integration Tests (Cross-Module)

Test workflows that span multiple services:

```bash
go test ./tests/integration/...
```

### 3. Contract Tests (Handler → Service)

Verify HTTP endpoints work correctly:

```bash
go test ./tests/contract/...
```

### 4. E2E Tests (Full User Flows)

Test complete user journeys:

- Registration → Browse opportunities → Register → Check-in → Log hours

## Next Steps

1. **Implement Service Adapters** (Highest Priority)

   - Create registration-opportunity adapter
   - Create hours-registration adapter
   - Create hours-volunteer adapter
   - Add FindVolunteersByOpportunity to registration repository

2. **Add Notification Service** (High Priority)

   - Centralized notification handling
   - Support in-app, email, and push notifications
   - Used by multiple modules

3. **Add Geocoding Service** (Medium Priority)

   - Address validation and geocoding
   - Optional but enhances location-based features

4. **Write Integration Tests** (High Priority)

   - Test cross-module workflows
   - Verify service adapters work correctly

5. **Run Full Test Suite** (Before Production)
   - Ensure all contracts pass
   - Verify no regressions
   - Check error handling

## Conclusion

✅ **All services are now properly initialized**
✅ **No nil pointer panics on startup**
✅ **Basic CRUD operations functional**
⚠️ **Cross-module workflows need adapters**
⚠️ **Notification system needs implementation**

The foundation is solid and the architecture is clean. The remaining work is to implement the cross-module communication patterns using adapter interfaces.
