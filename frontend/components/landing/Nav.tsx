"use client";

import Link from "next/link";
import { motion } from "framer-motion";
import { useUser, useStackApp } from "@stackframe/stack";

export function Nav() {
  const user = useUser();
  const stackApp = useStackApp();

  return (
    <nav className="sticky top-0 inset-x-0 z-50 bg-[#0d0d0d] border-b border-white/10">
      <div className="max-w-[1400px] mx-auto px-8 h-[72px] flex items-center justify-between">
        {/* Logo */}
        <Link href="/" className="flex items-center gap-2 hover:opacity-90 transition-opacity">
          <div className="relative flex-shrink-0 w-16 h-12">
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img src="/clawdface-logo.svg" alt="ClawdFace" className="w-full h-full object-contain" />
          </div>
          <span className="font-bold text-xl text-white">Clawd</span>
          <span className="font-bold text-xl text-red-500 -ml-1">Face</span>
        </Link>

        {/* Nav links */}
        <div className="hidden md:flex items-center gap-10 text-[15px] font-medium text-zinc-400">
          <Link href="/#how-it-works" className="hover:text-white transition-colors">How It Works?</Link>
          <Link href="/#features" className="hover:text-white transition-colors">Features</Link>
          <Link href="/#pricing" className="hover:text-white transition-colors">Pricing</Link>
          <Link href="/#contact" className="hover:text-white transition-colors">Contact Us</Link>
        </div>

        {/* Auth */}
        <div className="flex items-center gap-5">
          {user ? (
            <>
              <Link href="/dashboard">
                <motion.button
                  whileHover={{ scale: 1.03 }}
                  whileTap={{ scale: 0.97 }}
                  className="bg-[#00E3AA] text-black px-6 py-2.5 rounded-lg text-[15px] font-bold"
                >
                  Dashboard
                </motion.button>
              </Link>
              <button
                onClick={async () => {
                  try {
                    await stackApp.signOut();
                    window.location.href = "/";
                  } catch (e) {
                    console.error(e);
                  }
                }}
                className="text-[15px] font-medium text-zinc-400 hover:text-white transition-colors"
              >
                Log Out
              </button>
            </>
          ) : (
            <>
              <Link href="/log-in" className="text-[15px] font-medium text-zinc-300 hover:text-white transition-colors">
                Log In
              </Link>
              <Link href="/sign-up">
                <motion.button
                  whileHover={{ scale: 1.03 }}
                  whileTap={{ scale: 0.97 }}
                  className="bg-[#00E3AA] text-black px-6 py-2.5 rounded-lg text-[15px] font-bold"
                >
                  Sign Up
                </motion.button>
              </Link>
            </>
          )}
        </div>
      </div>
    </nav>
  );
}
