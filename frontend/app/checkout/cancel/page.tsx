"use client";

import { motion } from "framer-motion";
import Link from "next/link";

export default function PaymentCancelled() {
  return (
    <div className="min-h-screen bg-[#050505] flex items-center justify-center p-4 overflow-hidden relative">
      {/* Background glows */}
      <div className="pointer-events-none fixed inset-0 overflow-hidden">
        <div className="absolute top-1/3 left-1/2 -translate-x-1/2 w-[500px] h-[500px] bg-[#a78bfa]/4 rounded-full blur-[140px]" />
        <div className="absolute bottom-1/4 right-1/3 w-[300px] h-[300px] bg-red-500/3 rounded-full blur-[100px]" />
      </div>

      <div className="relative z-10 w-full max-w-lg">
        <motion.div
          initial={{ opacity: 0, y: 32, scale: 0.97 }}
          animate={{ opacity: 1, y: 0, scale: 1 }}
          transition={{ duration: 0.5, ease: "easeOut" }}
          className="bg-[#0f0f0f] border border-white/8 rounded-2xl p-10 flex flex-col items-center gap-7 text-center shadow-2xl"
        >
          {/* Icon */}
          <motion.div
            initial={{ scale: 0.6, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            transition={{ delay: 0.15, duration: 0.45, ease: "easeOut" }}
            className="w-20 h-20 rounded-full bg-white/5 border border-white/10 flex items-center justify-center"
          >
            <svg
              className="w-9 h-9 text-[#6b7280]"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              strokeWidth={2}
            >
              <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </motion.div>

          {/* Text */}
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.3, duration: 0.4 }}
            className="flex flex-col gap-2"
          >
            <span className="text-[11px] font-bold uppercase tracking-widest text-[#9ca3af] bg-white/5 px-3 py-1 rounded-full border border-white/8 self-center">
              Payment Cancelled
            </span>
            <h1 className="text-3xl font-bold text-white tracking-tight mt-1">
              No worries
            </h1>
            <p className="text-[#6b7280] text-[15px] leading-relaxed max-w-sm">
              You cancelled the checkout. Nothing was charged. You can upgrade whenever you&apos;re ready.
            </p>
          </motion.div>

          {/* Divider */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.45 }}
            className="w-full border-t border-white/5"
          />

          {/* Info row */}
          <motion.div
            initial={{ opacity: 0, y: 8 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.5, duration: 0.4 }}
            className="w-full flex flex-col gap-3"
          >
            <div className="flex items-start gap-3 text-left p-4 rounded-xl bg-white/[0.02] border border-white/5">
              <svg className="w-4 h-4 text-[#00E3AA] mt-0.5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <p className="text-sm text-[#9ca3af] leading-relaxed">
                Your current plan remains active. Head back to the subscription page to choose a plan when you&apos;re ready.
              </p>
            </div>
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
              <button className="w-full flex items-center justify-center gap-2 bg-[#00E3AA] text-black font-bold py-3.5 px-8 rounded-xl hover:bg-[#00E3AA]/90 active:scale-[0.98] transition-all duration-200 shadow-lg shadow-[#00E3AA]/20">
                View Plans
              </button>
            </Link>
            <Link href="/dashboard" className="w-full">
              <button className="w-full text-[#6b7280] hover:text-white text-sm py-2 transition-colors">
                ← Back to Dashboard
              </button>
            </Link>
          </motion.div>
        </motion.div>
      </div>
    </div>
  );
}
