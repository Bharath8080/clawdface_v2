"use client";

import { motion } from "framer-motion";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { Suspense } from "react";

function SuccessContent() {
  const searchParams = useSearchParams();
  const sessionId = searchParams.get("session_id");

  return (
    <div className="min-h-screen bg-canvas flex items-center justify-center p-4 overflow-hidden relative">
      {/* Background glows */}
      <div className="pointer-events-none fixed inset-0 overflow-hidden">
        <div className="absolute top-1/3 left-1/2 -translate-x-1/2 w-[600px] h-[600px] bg-brand/5 rounded-full blur-[140px]" />
        <div className="absolute bottom-1/4 right-1/3 w-[300px] h-[300px] bg-brand/3 rounded-full blur-[100px]" />
      </div>

      <div className="relative z-10 w-full max-w-lg">
        {/* Card */}
        <motion.div
          initial={{ opacity: 0, y: 32, scale: 0.97 }}
          animate={{ opacity: 1, y: 0, scale: 1 }}
          transition={{ duration: 0.5, ease: "easeOut" }}
          className="bg-[#0f0f0f] border border-white/8 rounded-2xl p-10 flex flex-col items-center gap-7 text-center shadow-2xl"
        >
          {/* Animated success ring */}
          <div className="relative">
            <motion.div
              initial={{ scale: 0.6, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              transition={{ delay: 0.15, duration: 0.5, ease: "easeOut" }}
              className="w-20 h-20 rounded-full bg-brand/10 border border-brand/25 flex items-center justify-center"
            >
              <motion.svg
                initial={{ pathLength: 0, opacity: 0 }}
                animate={{ pathLength: 1, opacity: 1 }}
                transition={{ delay: 0.4, duration: 0.5, ease: "easeOut" }}
                className="w-10 h-10 text-brand"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth={2.5}
              >
                <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
              </motion.svg>
            </motion.div>
            {/* Pulse ring */}
            <motion.div
              animate={{ scale: [1, 1.35, 1], opacity: [0.3, 0, 0.3] }}
              transition={{ duration: 2.5, repeat: Infinity, ease: "easeInOut" }}
              className="absolute inset-0 rounded-full border border-brand/30"
            />
          </div>

          {/* Text */}
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.3, duration: 0.4 }}
            className="flex flex-col gap-2"
          >
            <span className="text-[11px] font-bold uppercase tracking-widest text-brand bg-brand/10 px-3 py-1 rounded-full border border-brand/20 self-center">
              Payment Confirmed
            </span>
            <h1 className="text-3xl font-bold text-white tracking-tight mt-1">
              You&apos;re all set!
            </h1>
            <p className="text-[#6b7280] text-[15px] leading-relaxed max-w-sm">
              Your plan has been activated. Your AI avatars are ready — start a session from your dashboard.
            </p>
          </motion.div>

          {/* Divider */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.45 }}
            className="w-full border-t border-white/5"
          />

          {/* What's next checklist */}
          <motion.div
            initial={{ opacity: 0, y: 8 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.5, duration: 0.4 }}
            className="w-full flex flex-col gap-3 text-left"
          >
            {[
              "Plan activated on your account",
              "Session minutes are now available",
              "All avatar configurations unlocked",
            ].map((item, i) => (
              <div key={i} className="flex items-center gap-3">
                <div className="w-5 h-5 rounded-full bg-brand/15 border border-brand/30 flex items-center justify-center shrink-0">
                  <svg className="w-2.5 h-2.5 text-brand" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={3}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
                  </svg>
                </div>
                <span className="text-sm text-[#d1d5db]">{item}</span>
              </div>
            ))}
          </motion.div>

          {/* Divider */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.55 }}
            className="w-full border-t border-white/5"
          />

          {/* Actions */}
          <motion.div
            initial={{ opacity: 0, y: 8 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.6, duration: 0.4 }}
            className="w-full flex flex-col gap-3"
          >
            <Link href="/dashboard" className="w-full">
              <button className="w-full flex items-center justify-center gap-2 bg-brand text-black font-bold py-3.5 px-8 rounded-xl hover:bg-brand/90 active:scale-[0.98] transition-all duration-200 shadow-lg shadow-brand/20">
                <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
                </svg>
                Go to Dashboard
              </button>
            </Link>
            <Link href="/dashboard/settings/billing-and-subscription" className="w-full">
              <button className="w-full text-[#6b7280] hover:text-white text-sm py-2 transition-colors">
                View subscription details →
              </button>
            </Link>
          </motion.div>

          {/* Session ID (subtle, for reference) */}
          {sessionId && (
            <motion.p
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ delay: 0.8 }}
              className="text-[10px] text-[#374151] font-mono"
            >
              ref: {sessionId.slice(0, 24)}…
            </motion.p>
          )}
        </motion.div>
      </div>
    </div>
  );
}

export default function PaymentSuccess() {
  return (
    <Suspense>
      <SuccessContent />
    </Suspense>
  );
}
