import { supabase } from "./supabase";

/**
 * Persistent verified-users store backed by Supabase.
 * Survives server restarts and works in serverless environments.
 */

export async function isVerifiedUser(email: string): Promise<boolean> {
  try {
    const { data, error } = await supabase
      .from("verified_users")
      .select("email")
      .eq("email", email.toLowerCase())
      .single();

    if (error && error.code !== "PGRST116") { // PGRST116 is "no rows found"
      console.error("[userStore] Error checking verification:", error);
      return false;
    }

    return !!data;
  } catch (e) {
    console.error("[userStore] Unexpected error in isVerifiedUser:", e);
    return false;
  }
}

export async function registerVerifiedUser(email: string): Promise<void> {
  try {
    const { error } = await supabase
      .from("verified_users")
      .upsert({ email: email.toLowerCase() });

    if (error) {
      console.error("[userStore] Error registering user:", error);
    }
  } catch (e) {
    console.error("[userStore] Unexpected error in registerVerifiedUser:", e);
  }
}
