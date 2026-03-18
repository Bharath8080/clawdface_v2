import { db, profiles } from "@/lib/db";
import { eq } from "drizzle-orm";

/**
 * Persistent verified-users store using Supabase (via Drizzle) or Env Vars.
 * This ensures users stay verified even across Vercel deployments.
 */

export async function isVerifiedUser(email: string): Promise<boolean> {
  const cleanEmail = email.toLowerCase().trim();

  // 1. Check Database (Supabase)
  try {
    const user = await db.query.profiles.findFirst({
      where: eq(profiles.email, cleanEmail),
    });
    if (user) return true;
  } catch (e) {
    console.error("[userStore] DB check failed:", e);
  }

  // 2. Check Environment Variable (Best for quick whitelisting)
  const envEmails = process.env.VERIFIED_EMAILS || "";
  if (envEmails.split(",").map((e) => e.trim().toLowerCase()).includes(cleanEmail)) {
    return true;
  }

  return false;
}

export async function registerVerifiedUser(email: string): Promise<void> {
  const cleanEmail = email.toLowerCase().trim();

  try {
    await db.insert(profiles)
      .values({ email: cleanEmail })
      .onConflictDoUpdate({
        target: profiles.email,
        set: { email: cleanEmail },
      });
  } catch (e) {
    console.error("[userStore] DB register failed:", e);
  }
}
