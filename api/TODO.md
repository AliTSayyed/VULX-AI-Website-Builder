hexagonal DDD the typical order is:

Domain entities - Core business objects and rules
Application services - Use cases and orchestration logic
Ports (interfaces) - Define contracts for external dependencies
Handlers/Controllers - API layer that calls application services
Adapters - Concrete implementations of ports

1. Homepage (not logged in)
   ↓ [User clicks "Login with Google"]

2. Frontend calls beginAccountAuth()
   ↓ [Backend generates Google OAuth URL]

3. window.location.href = Google OAuth URL
   ↓ [Browser navigates to Google]

4. Google Consent Page
   ├─ [User clicks "Allow"] → Success path
   └─ [User clicks "Deny"] → Error path

5. Google redirects back to frontend
   ├─ Success: /auth/callback?code=ABC123&state=xyz
   └─ Error: /auth/callback?error=access_denied

6. AuthCallback component mounts
   ↓ [useEffect runs automatically]

7. useEffect checks URL parameters
   ├─ Has code → Call finishAccountAuth()
   └─ Has error/no code → navigate('/login')

8. Backend processes (success path only)
   ├─ Validates state
   ├─ Exchanges code for Google tokens  
   ├─ Gets user profile from Google
   ├─ Creates/updates user in database
   └─ Returns JWT + user profile

9. Frontend handles response
   ├─ Stores JWT in localStorage
   ├─ Updates React Query cache
   └─ navigate('/dashboard')

10. Final destination
    ├─ Success: Dashboard (authenticated)
    └─ Failure: Login page (try again)
