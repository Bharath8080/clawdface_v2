"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useUser } from "@stackframe/stack";
import Link from "next/link";

export default function EmailNotVerifiedPage() {
  const router = useRouter();
  const user = useUser();
  const contactChannels = user?.useContactChannels();
  const [isSubmitted, setIsSubmitted] = useState(false);
  const [sending, setSending] = useState(false);

  useEffect(() => {
    if (user === null) {
      router.replace("/log-in");
    } else if (user?.primaryEmailVerified) {
      router.replace("/dashboard");
    }
  }, [user, router]);

  const handleResend = async () => {
    const primaryChannel = contactChannels?.find(
      (ch) => ch.value === user?.primaryEmail
    );
    if (!primaryChannel) return;
    setSending(true);
    try {
      await primaryChannel.sendVerificationEmail({
        callbackUrl: `${window.location.origin}/verify-email`,
      });
      setIsSubmitted(true);
    } catch {
      // fail silently — button state already gives feedback
    } finally {
      setSending(false);
    }
  };

  if (isSubmitted) {
    return (
      <div className="min-h-screen bg-canvas flex items-center justify-center p-4">
        <div className="w-full max-w-md bg-surface-card border border-white/10 rounded-2xl p-8 flex flex-col items-center gap-6 text-center">
          <div className="w-16 h-16 bg-brand/10 border border-brand/20 rounded-full flex items-center justify-center">
            <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#00E3AA" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
              <path d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <h2 className="text-xl font-bold text-white">Email Sent</h2>
          <p className="text-zinc-400 text-sm leading-relaxed">
            A verification email has been sent to your inbox. Please check your
            email and follow the instructions to complete your registration.
          </p>
          <Link
            href="/log-in"
            className="mt-2 bg-brand text-black font-bold py-2.5 px-8 rounded-xl hover:brightness-110 transition-all"
          >
            Back to Login
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-canvas flex items-center justify-center p-4">
      <div className="w-full max-w-md bg-surface-card border border-white/10 rounded-2xl p-8 flex flex-col items-center gap-6 text-center">
        <div className="w-16 h-16 bg-red-500/10 border border-red-500/20 rounded-full flex items-center justify-center">
          <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#ef4444" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
            <circle cx="12" cy="12" r="10" />
            <path d="M15 9l-6 6M9 9l6 6" />
          </svg>
        </div>
        <div className="flex flex-col gap-2">
          <h2 className="text-xl font-bold text-white">Email Not Verified</h2>
          <p className="text-zinc-400 text-sm leading-relaxed">
            Please verify your email to access your account.
            {user?.primaryEmail && (
              <>
                {" "}We sent a link to{" "}
                <span className="text-white font-medium">{user.primaryEmail}</span>.
              </>
            )}
          </p>
        </div>
        <button
          onClick={handleResend}
          disabled={sending}
          className="w-full bg-brand text-black font-bold py-2.5 rounded-xl hover:brightness-110 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {sending ? "Sending..." : "Send Verification Email"}
        </button>
        <button
          onClick={() => user?.signOut()}
          className="text-sm text-zinc-500 hover:text-zinc-300 transition-colors"
        >
          Sign out
        </button>
      </div>
    </div>
  );
}
