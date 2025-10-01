hexagonal DDD the typical order is:

Domain entities - Core business objects and rules
Application services - Use cases and orchestration logic
Ports (interfaces) - Define contracts for external dependencies
Handlers/Controllers - API layer that calls application services
Adapters - Concrete implementations of ports

## Authentication/Authorization Flow (Concise)

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
   ├─ Browser stores JWT cookie automatically
   ├─ Updates React Query cache
   └─ navigate('/dashboard')

10. Final destination
    ├─ Success: Dashboard (authenticated)
    └─ Failure: Login page (try again)

## JWT Flow Standard (10,000 + users and high security needed)

**Login:**

- Frontend → credentials → Server
- Server → validates → returns access token (15min) + refresh token (7days)
- Frontend → stores both tokens

**API Call:**

- Frontend → sends access token → Server
- Server → validates → returns data

**Access Token Expired:**

- Frontend → sends expired access token → Server
- Server → returns 401
- Frontend interceptor → calls `/refresh` with refresh token → Server
- Server → validates refresh token → returns NEW access + NEW refresh tokens → invalidates old refresh
- Frontend → retries original request with new access token → Server
- Server → returns data

**Refresh Token Expired:**

- Frontend → calls `/refresh` → Server
- Server → returns 401/403
- Frontend → clears tokens → redirects to login

## JWT Flow Hybrid (100 + users and less security needed)

**Login/Signup:**

- Frontend → credentials → Server
- Server → validates → creates 7-day JWT → returns in httpOnly cookie
- Browser → stores cookie automatically

**Normal Request:**

- Frontend → API call → Browser attaches cookie → Server
- Server → validates JWT → processes request

**Token Expiring Soon (<42 hours left):**

- Frontend → API call → Server
- Server → validates JWT → detects <42 hours left → processes request
- Server → generates NEW 7-day JWT → adds to response `Set-Cookie` header
- Browser → auto-updates cookie with new token

**Token Fully Expired (>7 days inactive):**

- Frontend → API call → Server
- Server → returns `Unauthenticated` (401)
- Frontend interceptor → clears cache → redirects to login

**Logout:**

- Frontend → calls logout → Server
- Server → adds hashed token to blacklist DB (until natural expiry)
- Server → clears cookie via `Set-Cookie` with past expiry
- Browser → deletes cookie
- Frontend → clears cache → redirects to login

**Key:** Single long-lived token (7 days), auto-refreshes before expiry, httpOnly cookies, DB blacklist for revocation.

**Perfect summary. Yes, exactly.**

## Your App's Token Strategy:

**No separate refresh token** - just one auth token that:

1. **Auto-refreshes proactively** (when <42 hours left) → user stays logged in seamlessly
2. **Expires after 7 days of inactivity** → user logged out, must re-login
3. **Blacklisted on manual logout** → token revoked immediately, can't be reused

**Two ways a token becomes invalid:**

1. **Natural expiry** (7 days pass, not refreshed) → `Unauthenticated` → login page
2. **Manual logout** (added to blacklist) → `Unauthenticated` → login page

**Key insight:** The "refresh" happens invisibly on the backend during normal requests, not via a separate frontend flow. User never knows it's happening.
