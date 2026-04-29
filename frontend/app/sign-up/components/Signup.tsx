"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useStackApp } from "@stackframe/stack";

function isValidEmail(email: string) {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
}

function isValidPassword(password: string) {
  return /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,}$/.test(password);
}


const Signup = () => {
  const router = useRouter();
  const app = useStackApp();

  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [companyName, setCompanyName] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [acceptedTerms, setAcceptedTerms] = useState(false);
  const [acceptedNews, setAcceptedNews] = useState(false);

  const [nameError, setNameError] = useState("");
  const [emailError, setEmailError] = useState("");
  const [companyNameError, setCompanyNameError] = useState("");
  const [passwordError, setPasswordError] = useState("");
  const [confirmPasswordError, setConfirmPasswordError] = useState("");
  const [formError, setFormError] = useState("");

  const [loading, setLoading] = useState(false);
  const [isSubmitted, setIsSubmitted] = useState(false);

  useEffect(() => {
    const timer = setTimeout(() => setNameError(""), 5000);
    return () => clearTimeout(timer);
  }, [nameError]);

  useEffect(() => {
    const timer = setTimeout(() => setEmailError(""), 5000);
    return () => clearTimeout(timer);
  }, [emailError]);

  useEffect(() => {
    const timer = setTimeout(() => setPasswordError(""), 5000);
    return () => clearTimeout(timer);
  }, [passwordError]);

  useEffect(() => {
    const timer = setTimeout(() => setConfirmPasswordError(""), 5000);
    return () => clearTimeout(timer);
  }, [confirmPasswordError]);

  const handleSignUp = useCallback(async () => {
    setNameError("");
    setEmailError("");
    setCompanyNameError("");
    setPasswordError("");
    setConfirmPasswordError("");
    setFormError("");

    if (!name) { setNameError("Name is required"); return; }
    if (!email) { setEmailError("Email is required"); return; }
    if (!companyName) { setCompanyNameError("Company is required"); return; }
    if (!password) { setPasswordError("Password is required"); return; }
    if (!confirmPassword) { setConfirmPasswordError("Confirm password is required"); return; }

    if (!acceptedTerms) {
      setFormError("Please accept the Terms of use and Privacy Policy");
      return;
    }
    if (!isValidEmail(email)) { setEmailError("Please enter a valid email"); return; }
    if (password !== confirmPassword) { setConfirmPasswordError("Passwords do not match"); return; }
    if (!isValidPassword(password)) {
      setPasswordError("Password must contain uppercase, lowercase, number, and one special character from @$!%*?&");
      return;
    }

    setLoading(true);
    const result = await app.signUpWithCredential({
      email,
      password,
      noRedirect: true,
      verificationCallbackUrl: `${window.location.origin}/verify-email`,
    });

    console.log("Sign up result:", result);

    if (result.status === "ok") {
      // User exists in Stack Auth but session isn't active until email is verified.
      // Backend user creation happens in the dashboard initData once they log in.
      setIsSubmitted(true);
    }

    if ((result as any)?.error?.name === "KnownError<USER_EMAIL_ALREADY_EXISTS>") {
      setEmailError("Email already exists");
    } else if (result?.status === "error") {
      setFormError(result?.error?.message ?? "Sign up failed");
    }

    setLoading(false);
  }, [app, name, email, companyName, password, confirmPassword, acceptedTerms, acceptedNews]);

  useEffect(() => {
    const handler = (e: KeyboardEvent) => { if (e.key === "Enter") handleSignUp(); };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [handleSignUp]);

  if (isSubmitted) {
    return (
      <div className="min-h-screen bg-canvas flex items-center justify-center p-4">
        <div className="w-full max-w-md bg-surface-card border border-white/10 rounded-2xl p-8 flex flex-col items-center gap-6 text-center">
          <div className="w-16 h-16 bg-brand/10 rounded-full flex items-center justify-center">
            <svg xmlns="http://www.w3.org/2000/svg" className="h-8 w-8 text-brand" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <h2 className="text-2xl font-bold text-white">Check your email</h2>
          <p className="text-zinc-400 text-sm">
            A verification email has been sent to{" "}
            <span className="text-white font-medium">{email}</span>.
            Please verify your email to complete registration.
          </p>
          <button
            onClick={() => router.replace("/log-in")}
            className="mt-2 bg-brand text-black font-bold py-3 px-8 rounded-xl hover:bg-brand/90 transition-colors"
          >
            Back to Login
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-canvas flex items-center justify-center p-4">
      <div className="w-full max-w-md bg-surface-card border border-white/10 rounded-2xl p-8 flex flex-col gap-5">
        <div className="flex flex-col items-center gap-2 mb-1">
          <span className="text-2xl font-bold text-white tracking-tight">ClawdFace</span>
          <span className="text-sm text-zinc-400">Create your account</span>
        </div>

        {formError && (
          <div className="bg-red-500/10 border border-red-500/30 text-red-400 text-sm rounded-lg px-4 py-3">
            {formError}
          </div>
        )}

        <div className="flex flex-col gap-1">
          <label className="text-sm font-medium text-zinc-300">Name</label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Enter your name"
            className="bg-surface-elevated border border-white/10 rounded-xl px-4 py-3 text-white placeholder-zinc-600 focus:outline-none focus:border-brand/50 transition-colors"
          />
          {nameError && <span className="text-red-400 text-xs">{nameError}</span>}
        </div>

        <div className="flex flex-col gap-1">
          <label className="text-sm font-medium text-zinc-300">Email</label>
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="Enter your email"
            className="bg-surface-elevated border border-white/10 rounded-xl px-4 py-3 text-white placeholder-zinc-600 focus:outline-none focus:border-brand/50 transition-colors"
          />
          {emailError && <span className="text-red-400 text-xs">{emailError}</span>}
        </div>

        <div className="flex flex-col gap-1">
          <label className="text-sm font-medium text-zinc-300">Company</label>
          <input
            type="text"
            value={companyName}
            onChange={(e) => setCompanyName(e.target.value)}
            placeholder="Enter company name"
            className="bg-surface-elevated border border-white/10 rounded-xl px-4 py-3 text-white placeholder-zinc-600 focus:outline-none focus:border-brand/50 transition-colors"
          />
          {companyNameError && <span className="text-red-400 text-xs">{companyNameError}</span>}
        </div>

        <div className="flex flex-col gap-1">
          <label className="text-sm font-medium text-zinc-300">Password</label>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="Enter password"
            className="bg-surface-elevated border border-white/10 rounded-xl px-4 py-3 text-white placeholder-zinc-600 focus:outline-none focus:border-brand/50 transition-colors"
          />
          {passwordError && <span className="text-red-400 text-xs">{passwordError}</span>}
        </div>

        <div className="flex flex-col gap-1">
          <label className="text-sm font-medium text-zinc-300">Confirm Password</label>
          <input
            type="password"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            placeholder="Confirm your password"
            className="bg-surface-elevated border border-white/10 rounded-xl px-4 py-3 text-white placeholder-zinc-600 focus:outline-none focus:border-brand/50 transition-colors"
          />
          {confirmPasswordError && <span className="text-red-400 text-xs">{confirmPasswordError}</span>}
        </div>

        <div className="flex flex-col gap-2">
          <label className="flex items-start gap-3 cursor-pointer">
            <input
              type="checkbox"
              checked={acceptedTerms}
              onChange={(e) => setAcceptedTerms(e.target.checked)}
              className="mt-0.5 accent-brand"
            />
            <span className="text-sm text-zinc-400">
              I have read and accepted the{" "}
              <Link href="/terms-of-service" target="_blank" className="text-brand hover:underline">Terms of use</Link>
              {" "}and{" "}
              <Link href="/privacy-policy" target="_blank" className="text-brand hover:underline">Privacy Policy</Link>
            </span>
          </label>
          <label className="flex items-start gap-3 cursor-pointer">
            <input
              type="checkbox"
              checked={acceptedNews}
              onChange={(e) => setAcceptedNews(e.target.checked)}
              className="mt-0.5 accent-brand"
            />
            <span className="text-sm text-zinc-400">
              Keep me informed about products and services
            </span>
          </label>
        </div>

        <button
          onClick={handleSignUp}
          disabled={loading || !acceptedTerms}
          className="w-full bg-brand text-black font-bold py-3 rounded-xl hover:bg-brand/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {loading ? "Creating account..." : "Sign Up"}
        </button>

        <span className="text-sm text-zinc-400 text-center">
          Already have an account?{" "}
          <Link href="/log-in" className="text-brand hover:underline font-medium">
            Log in
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

export default Signup;
