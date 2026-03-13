# ClawdFace Frontend: Real-time Interactive UI

This directory contains the production-grade frontend for the ClawdFace platform, built with **Next.js 15** and **LiveKit Components**. It is designed for maximum speed, security, and responsiveness.

---

## 🏗️ Core Architecture: Drizzle ORM & Server Actions

ClawdFace now uses **Drizzle ORM** for persistent state management, ensuring data integrity for profiles, bots, and conversation history.

1.  **Server Actions (Primary)**: All database operations are encapsulated in `lib/database-actions.ts`. These are Next.js Server Actions (`"use server"`) that execute on the server, keeping your Supabase credentials secure and away from the browser.
2.  **Browser-Safe Client**: A dedicated Supabase client in `lib/supabase-client.ts` is used for client-side authentication while remaining isolated from server-only logic.
3.  **Supabase Shared Pooler**: Optimized database connections via the Supabase Shared Pooler, enabling efficient scaling for serverless deployments.

---

## 🔐 Authentication & Verification

- **Integration**: Powered by `@react-oauth/google` using the **Implicit Flow**.
- **Context Provider**: `GoogleOAuthProvider` wraps the application (see `components/Providers.tsx`).
- **Hook-based Login**: Uses the `useGoogleLogin` hook to fetch high-fidelity profiles from the Google UserInfo API.
- **Persistence**: Successful sign-ins are stored in `localStorage` under `clawdface_auth`.
- **Environment Verification**: Access is restricted using the `VERIFIED_EMAILS` environment variable. This allows the administrator to authorize users globally via environment settings without managing a database table.

---

## 🎨 UI/UX Philosophy

- **Glassmorphism**: High-transparency layers with backdrop blurs.
- **Framer Motion**: Utilized for "Spring" physics based transitions between disconnected and connected states.
- **Real-time Visualization**: Custom `BarVisualizer` integration for high-fidelity audio feedback.
- **Responsive Sidebar**: Collapsible navigation with active session tracking.

---

## 📂 Directory Structure

| Path | Purpose |
| :--- | :--- |
| `app/api/` | Serverless endpoints (Connection details, OTP logic). |
| `drizzle/` | **[NEW]** Drizzle schema and client initialization. |
| `components/` | Reusable React components (Sidebar, ControlBar, Visualizers). |
| `lib/database-actions.ts` | **[NEW]** Server-side database operations. |
| `lib/supabase-client.ts` | **[NEW]** Browser-safe Supabase client. |
| `lib/auth.ts` | Client-side auth utilities and local storage management. |
| `public/` | Assets and brand identifiers. |

---

## 🔧 Frontend API Routes

### `POST /api/connection-details`
Generates an ephemeral LiveKit Token.
- **Feature**: Automatically embeds the user's OpenClaw configuration into the **Participant Metadata**. This is what allows the agent to be stateless.

### `POST /api/user-config`
Handles the sync between the browser and the local filesystem.

### `POST /api/send-otp` / `verify-otp`
Simple email-based verification system for authorized users.

---

