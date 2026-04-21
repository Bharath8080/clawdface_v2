"use client";

import { useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { useStackApp } from "@stackframe/stack";
import Link from "next/link";

export default function VerifyEmailPage() {
  const params = useSearchParams();
  const router = useRouter();
  const app = useStackApp();
  const [status, setStatus] = useState<"loading" | "success" | "error">("loading");

  useEffect(() => {
    const code = params.get("code");
    if (!code) { setStatus("error"); return; }

    app.verifyEmail(code).then((result) => {
      if (result.status === "ok") {
        setStatus("success");
        setTimeout(() => router.push("/log-in"), 2500);
      } else {
        setStatus("error");
      }
    }).catch(() => setStatus("error"));
  }, [params, app, router]);

  return (
    <div className="min-h-screen bg-[#050505] flex items-center justify-center p-4">
      <div className="w-full max-w-md bg-[#111111] border border-white/10 rounded-2xl p-8 flex flex-col items-center gap-6 text-center">

        {status === "loading" && (
          <>
            <svg className="animate-spin" width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="#00E3AA" strokeWidth="2">
              <circle cx="12" cy="12" r="10" opacity="0.25" />
              <path d="M22 12a10 10 0 0 1-10 10" opacity="0.9" />
            </svg>
            <h2 className="text-xl font-bold text-white">Verifying Email</h2>
            <p className="text-zinc-400 text-sm">Please wait while your email is being verified&hellip;</p>
          </>
        )}

        {status === "success" && (
          <>
            <div className="w-16 h-16 bg-[#00E3AA]/10 border border-[#00E3AA]/20 rounded-full flex items-center justify-center">
              <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#00E3AA" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
                <path d="M5 13l4 4L19 7" />
              </svg>
            </div>
            <h2 className="text-xl font-bold text-white">Email Verified</h2>
            <p className="text-zinc-400 text-sm">
              Your email has been verified. You will be redirected to login shortly.
            </p>
            <Link
              href="/log-in"
              className="mt-2 bg-[#00E3AA] text-black font-bold py-2.5 px-8 rounded-xl hover:brightness-110 transition-all"
            >
              Sign In
            </Link>
          </>
        )}

        {status === "error" && (
          <>
            <div className="w-16 h-16 bg-red-500/10 border border-red-500/20 rounded-full flex items-center justify-center">
              <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#ef4444" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
                <circle cx="12" cy="12" r="10" />
                <path d="M15 9l-6 6M9 9l6 6" />
              </svg>
            </div>
            <h2 className="text-xl font-bold text-white">Invalid or expired link</h2>
            <p className="text-zinc-400 text-sm">
              This verification link is no longer valid. Please request a new one.
            </p>
            <Link
              href="/email-not-verified"
              className="w-full bg-[#00E3AA] text-black font-bold py-2.5 rounded-xl hover:brightness-110 transition-all text-center"
            >
              Resend Verification Email
            </Link>
          </>
        )}

      </div>
    </div>
  );
}
