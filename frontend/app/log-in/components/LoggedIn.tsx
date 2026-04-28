"use client";

import Link from "next/link";
import React, { useCallback, useEffect, useRef, useState } from "react";
import { useSearchParams } from "next/navigation";
import { useStackApp } from "@stackframe/stack";

function isValidEmail(email: string) {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
}

const LoggedIn = () => {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [emailError, setEmailError] = useState("");
  const [passwordError, setPasswordError] = useState("");
  const [formError, setFormError] = useState("");
  const [loading, setLoading] = useState(false);

  const searchParams = useSearchParams();
  const redirectTo = searchParams.get("redirectTo");

  useEffect(() => {
    if (redirectTo) {
      localStorage.setItem("redirectTo", redirectTo);
    }
  }, [redirectTo]);

  const logInRef = useRef<HTMLDivElement>(null);
  const app = useStackApp();

  const handleLogin = useCallback(async () => {
    setEmailError("");
    setPasswordError("");
    setFormError("");

    if (!email) {
      setEmailError("Email is required");
      return;
    }
    if (!password) {
      setPasswordError("Password is required");
      return;
    }
    if (!isValidEmail(email)) {
      setEmailError("Please enter a valid email");
      return;
    }

    setLoading(true);
    const result = await app?.signInWithCredential({ email, password });
    if (result?.status === "error") {
      setFormError(result.error.message);
    }
    setLoading(false);
  }, [app, email, password]);

  useEffect(() => {
    let timer: NodeJS.Timeout | null = null;
    if (emailError || passwordError || formError) {
      timer = setTimeout(() => {
        setEmailError("");
        setPasswordError("");
        setFormError("");
      }, 5000);
    }
    return () => { if (timer) clearTimeout(timer); };
  }, [emailError, passwordError, formError]);

  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      if (e.key === "Enter") handleLogin();
    },
    [handleLogin],
  );

  useEffect(() => {
    const el = logInRef.current;
    if (el) el.addEventListener("keydown", handleKeyDown);
    return () => { if (el) el.removeEventListener("keydown", handleKeyDown); };
  }, [handleKeyDown]);

  return (
    <div className="min-h-screen bg-canvas flex items-center justify-center p-4">
      <div
        ref={logInRef}
        className="w-full max-w-md bg-surface-card border border-white/10 rounded-2xl p-8 flex flex-col gap-6"
      >
        <div className="flex flex-col items-center gap-2 mb-2">
          <span className="text-2xl font-bold text-white tracking-tight">ClawdFace</span>
          <span className="text-sm text-zinc-400">Sign in to your account</span>
        </div>

        {formError && (
          <div className="bg-red-500/10 border border-red-500/30 text-red-400 text-sm rounded-lg px-4 py-3">
            {formError}
          </div>
        )}

        <div className="flex flex-col gap-1">
          <label className="text-sm font-medium text-zinc-300">Email</label>
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="Enter your email"
            className="bg-surface-elevated border border-white/10 rounded-xl px-4 py-3 text-white placeholder-zinc-600 focus:outline-none focus:border-brand/50 transition-colors"
          />
          {emailError && <span className="text-red-400 text-xs mt-1">{emailError}</span>}
        </div>

        <div className="flex flex-col gap-1">
          <label className="text-sm font-medium text-zinc-300">Password</label>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="Enter your password"
            className="bg-surface-elevated border border-white/10 rounded-xl px-4 py-3 text-white placeholder-zinc-600 focus:outline-none focus:border-brand/50 transition-colors"
          />
          {passwordError && <span className="text-red-400 text-xs mt-1">{passwordError}</span>}
        </div>

        <Link
          href="/reset-password"
          className="text-sm text-brand hover:underline self-start -mt-3"
        >
          Forgot your password?
        </Link>

        <button
          onClick={handleLogin}
          disabled={loading}
          className="w-full bg-brand text-black font-bold py-3 rounded-xl hover:bg-brand/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {loading ? "Signing in..." : "Sign In"}
        </button>

        <span className="text-sm text-zinc-400 text-center">
          Don&apos;t have an account?{" "}
          <Link href="/sign-up" className="text-brand hover:underline font-medium">
            Sign Up
          </Link>
        </span>

        <div className="flex items-center gap-3">
          <div className="flex-1 h-px bg-white/10" />
          <span className="text-xs text-zinc-500">OR</span>
          <div className="flex-1 h-px bg-white/10" />
        </div>

        <button
          onClick={async () => { await app.signInWithOAuth("google"); }}
          className="w-full flex items-center justify-center gap-3 bg-surface-elevated border border-white/10 text-white font-medium py-3 rounded-xl hover:bg-white/5 transition-colors"
        >
          <svg className="w-5 h-5" viewBox="0 0 24 24">
            <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" />
            <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" />
            <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" />
            <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" />
          </svg>
          Google
        </button>
      </div>
    </div>
  );
};

export default LoggedIn;
