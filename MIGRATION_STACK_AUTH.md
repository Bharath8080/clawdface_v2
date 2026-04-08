# Stack Auth Migration Summary

This document outlines the architectural changes, file modifications, and migration steps taken to transition the `ClawdFace` application from the custom Google OAuth implementation to **Stack Auth**.

## Overview
The application previously relied heavily on a client-side authentication model leveraging `@react-oauth/google` and `localStorage` state management (`lib/auth.ts`).

To leverage a production-ready authentication infrastructure, we migrated to **Stack Auth**, which provides robust cookie-based server-side session management natively designed for Next.js 14/15 App Router applications.

---

## 📦 Dependency Changes
- **Added**: `@stackframe/stack` (Installed via `npm install @stackframe/stack --legacy-peer-deps`)
- **Deprecated**: `@react-oauth/google` (No longer in use; safe to uninstall)

---

## 📁 File Structure Changes

### Files Added
1. `frontend/stack.ts`
   - **Purpose**: Creates the `stackServerApp` object using the `nextjs-cookie` token store. It uses `"server-only"` ensuring that sensitive server tokens never bleed into the client.
2. `frontend/app/handler/[...stack]/page.tsx`
   - **Purpose**: The dynamic Catch-All API Route handler required by Stack Auth to power isolated login, sign-up, and account management views automatically.

### Files Removed
1. `frontend/app/login/page.tsx`
2. `frontend/app/signup/page.tsx`
   - **Reason**: The custom-built Google Login/Signup views were deleted as we now securely offload all authentication screens and UI routing directly to Stack Auth's built-in component routes (`/handler/sign-in`, `/handler/sign-up`).

---

## 📝 Files Modified

### 1. `frontend/.env.local`
- **What Changed**: Appended the project keys necessary to link the frontend to the Stack Auth cloud.
- **Additions**:
  ```env
  NEXT_PUBLIC_STACK_PROJECT_ID=0a4ade1b-9028-4482-948d-dea78677f008
  STACK_SECRET_SERVER_KEY=ssk_qptva...
  ```

### 2. `frontend/app/layout.tsx`
- **What Changed**: Wrapped the entire DOM structure within the Next.js `RootLayout`.
- **Additions**: Added `<StackProvider app={stackServerApp}>` and `<StackTheme>` wrapping over the `<Providers>` chunk. This powers the hooks for all client components while authenticating server elements efficiently.

### 3. `frontend/components/Providers.tsx`
- **What Changed**: Stripped out legacy Google Auth configurations.
- **Deletions**: Removed `<GoogleOAuthProvider>`. This file is now exclusively used to initialize client-side DOM theme listeners (Light/Dark mode) upon hydration.

### 4. `frontend/lib/auth.ts`
- **What Changed**: Repurposed from an Authentication engine to a lightweight Theme & Utility helper.
- **Deletions**: Removed `getUser()`, `isAuthenticated()`, `saveUserToLocalStorage()`, `logout()`, and the `AuthUser` interface.
- **Additions**: Modified `getInitials(email, name)` to accept optional string arguments directly extracted from the Stack Auth User object.

### 5. `frontend/app/page.tsx`
- **What Changed**: Replaced localized localStorage polling with reactive Stack hooks.
- **Modifications**: 
  - Substituted the static `getUser()` fetch with `@stackframe/stack`'s `useUser()` hook.
  - Authentication checks within the `useEffect` closure now look for `user === null` and immediately execute `router.replace("/handler/sign-in")`.
  - Re-mapped properties like `user?.email` to Stack's standard `user.primaryEmail || user.displayName` before executing database synchronization actions (`syncUserAction`, `fetchBots`, `createConversationAction`).

### 6. `frontend/components/Sidebar.tsx`
- **What Changed**: Extracted user metadata correctly from Stack context.
- **Modifications**:
  - Implemented `useUser` and `useStackApp` hooks.
  - The "Sign Out" flow now correctly executes the asynchronous `await app.signOut()` command ensuring secure invalidation of cookies, followed by a redirect to `/handler/sign-in`.
  - Replaced `user.picture` property paths with `user.profileImageUrl` alongside dynamic fallback initial rendering.

---

## 🛠️ Post-Migration Next Steps
- Start the server using `npm run dev`. Navigate to the application without a session, and you should be immediately redirected to the `/handler/sign-in` screen powered by Stack Auth.
- Create your first user account via Stack Auth, verify the primary email is correctly tracked, and test completing a `DirectCall` or `bot` save event to ensure the `supabase` instance syncs your email properly.
