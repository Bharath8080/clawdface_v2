"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";
import { useStackApp } from "@stackframe/stack";

function isValidEmail(email: string) {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
}

export default function ResetPasswordPage() {
  const app = useStackApp();
  const searchParams = useSearchParams();
  const code = searchParams.get("code");

  // ── Forgot-password state ────────────────────────────────────────────────
  const [email, setEmail] = useState("");
  const [emailError, setEmailError] = useState("");
  const [sendLoading, setSendLoading] = useState(false);
  const [emailSent, setEmailSent] = useState(false);

  // ── Set-new-password state ───────────────────────────────────────────────
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [passwordError, setPasswordError] = useState("");
  const [resetLoading, setResetLoading] = useState(false);
  const [resetDone, setResetDone] = useState(false);
  const [formError, setFormError] = useState("");

  useEffect(() => {
    if (emailError) {
      const t = setTimeout(() => setEmailError(""), 5000);
      return () => clearTimeout(t);
    }
  }, [emailError]);

  useEffect(() => {
    if (passwordError || formError) {
      const t = setTimeout(() => { setPasswordError(""); setFormError(""); }, 5000);
      return () => clearTimeout(t);
    }
  }, [passwordError, formError]);

  const handleSendEmail = useCallback(async () => {
    if (!email) { setEmailError("Email is required"); return; }
    if (!isValidEmail(email)) { setEmailError("Please enter a valid email"); return; }
    setSendLoading(true);
    try {
      await app.sendForgotPasswordEmail(email);
      setEmailSent(true);
    } catch (err: unknown) {
      setEmailError(err instanceof Error ? err.message : "Failed to send reset email");
    } finally {
      setSendLoading(false);
    }
  }, [app, email]);

  const handleReset = useCallback(async () => {
    if (!password) { setPasswordError("Password is required"); return; }
    if (password !== confirmPassword) { setPasswordError("Passwords do not match"); return; }
    if (!code) return;
    setResetLoading(true);
    try {
      await app.resetPassword({ password, code });
      setResetDone(true);
    } catch (err: unknown) {
      setFormError(err instanceof Error ? err.message : "Failed to reset password. The link may have expired.");
    } finally {
      setResetLoading(false);
    }
  }, [app, password, confirmPassword, code]);

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.key !== "Enter") return;
      if (code) handleReset();
      else handleSendEmail();
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [code, handleReset, handleSendEmail]);

  // ── Success: email sent ──────────────────────────────────────────────────
  if (emailSent) {
    return (
      <div className="min-h-screen bg-[#050505] flex items-center justify-center p-4">
        <div className="w-full max-w-md bg-[#111111] border border-white/10 rounded-2xl p-8 flex flex-col items-center gap-6 text-center">
          <div className="w-16 h-16 bg-[#00E3AA]/10 rounded-full flex items-center justify-center">
            <svg className="h-8 w-8 text-[#00E3AA]" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
            </svg>
          </div>
          <h2 className="text-2xl font-bold text-white">Check your email</h2>
          <p className="text-zinc-400 text-sm">
            We sent a password reset link to{" "}
            <span className="text-white font-medium">{email}</span>.
            Check your inbox and click the link to reset your password.
          </p>
          <Link
            href="/log-in"
            className="mt-2 bg-[#00E3AA] text-black font-bold py-3 px-8 rounded-xl hover:bg-[#00E3AA]/90 transition-colors"
          >
            Back to Login
          </Link>
        </div>
      </div>
    );
  }

  // ── Success: password reset ──────────────────────────────────────────────
  if (resetDone) {
    return (
      <div className="min-h-screen bg-[#050505] flex items-center justify-center p-4">
        <div className="w-full max-w-md bg-[#111111] border border-white/10 rounded-2xl p-8 flex flex-col items-center gap-6 text-center">
          <div className="w-16 h-16 bg-[#00E3AA]/10 rounded-full flex items-center justify-center">
            <svg className="h-8 w-8 text-[#00E3AA]" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <h2 className="text-2xl font-bold text-white">Password updated</h2>
          <p className="text-zinc-400 text-sm">
            Your password has been reset successfully. You can now sign in with your new password.
          </p>
          <Link
            href="/log-in"
            className="mt-2 bg-[#00E3AA] text-black font-bold py-3 px-8 rounded-xl hover:bg-[#00E3AA]/90 transition-colors"
          >
            Sign In
          </Link>
        </div>
      </div>
    );
  }

  // ── Set new password (came from email link) ──────────────────────────────
  if (code) {
    return (
      <div className="min-h-screen bg-[#050505] flex items-center justify-center p-4">
        <div className="w-full max-w-md bg-[#111111] border border-white/10 rounded-2xl p-8 flex flex-col gap-6">
          <div className="flex flex-col items-center gap-2 mb-2">
            <span className="text-2xl font-bold text-white tracking-tight">ClawdFace</span>
            <span className="text-sm text-zinc-400">Set your new password</span>
          </div>

          {formError && (
            <div className="bg-red-500/10 border border-red-500/30 text-red-400 text-sm rounded-lg px-4 py-3">
              {formError}
            </div>
          )}

          <div className="flex flex-col gap-1">
            <label className="text-sm font-medium text-zinc-300">New Password</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Enter new password"
              className="bg-[#1a1a1a] border border-white/10 rounded-xl px-4 py-3 text-white placeholder-zinc-600 focus:outline-none focus:border-[#00E3AA]/50 transition-colors"
            />
            {passwordError && <span className="text-red-400 text-xs mt-1">{passwordError}</span>}
          </div>

          <div className="flex flex-col gap-1">
            <label className="text-sm font-medium text-zinc-300">Confirm Password</label>
            <input
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              placeholder="Confirm new password"
              className="bg-[#1a1a1a] border border-white/10 rounded-xl px-4 py-3 text-white placeholder-zinc-600 focus:outline-none focus:border-[#00E3AA]/50 transition-colors"
            />
          </div>

          <button
            onClick={handleReset}
            disabled={resetLoading}
            className="w-full bg-[#00E3AA] text-black font-bold py-3 rounded-xl hover:bg-[#00E3AA]/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {resetLoading ? "Updating..." : "Set New Password"}
          </button>

          <span className="text-sm text-zinc-400 text-center">
            <Link href="/log-in" className="text-[#00E3AA] hover:underline font-medium">
              Back to Login
            </Link>
          </span>
        </div>
      </div>
    );
  }

  // ── Forgot password (enter email) ────────────────────────────────────────
  return (
    <div className="min-h-screen bg-[#050505] flex items-center justify-center p-4">
      <div className="w-full max-w-md bg-[#111111] border border-white/10 rounded-2xl p-8 flex flex-col gap-6">
        <div className="flex flex-col items-center gap-2 mb-2">
          <span className="text-2xl font-bold text-white tracking-tight">ClawdFace</span>
          <span className="text-sm text-zinc-400">Reset your password</span>
        </div>

        <p className="text-zinc-400 text-sm text-center -mt-2">
          Enter your email and we&apos;ll send you a link to reset your password.
        </p>

        <div className="flex flex-col gap-1">
          <label className="text-sm font-medium text-zinc-300">Email</label>
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="Enter your email"
            className="bg-[#1a1a1a] border border-white/10 rounded-xl px-4 py-3 text-white placeholder-zinc-600 focus:outline-none focus:border-[#00E3AA]/50 transition-colors"
          />
          {emailError && <span className="text-red-400 text-xs mt-1">{emailError}</span>}
        </div>

        <button
          onClick={handleSendEmail}
          disabled={sendLoading}
          className="w-full bg-[#00E3AA] text-black font-bold py-3 rounded-xl hover:bg-[#00E3AA]/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {sendLoading ? "Sending..." : "Send Reset Link"}
        </button>

        <span className="text-sm text-zinc-400 text-center">
          Remember your password?{" "}
          <Link href="/log-in" className="text-[#00E3AA] hover:underline font-medium">
            Sign In
          </Link>
        </span>
      </div>
    </div>
  );
}
