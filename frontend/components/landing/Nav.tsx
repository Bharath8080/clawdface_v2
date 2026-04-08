"use client";

import Link from "next/link";
import Image from "next/image";
import { motion } from "framer-motion";
import { useUser, useStackApp } from "@stackframe/stack";

export function Nav() {
  const user = useUser();
  const stackApp = useStackApp();
  return (
    <nav className="sticky top-0 inset-x-0 z-50 bg-[#050505]/80 backdrop-blur-xl border-b border-white/5">
      <div className="max-w-[1400px] mx-auto px-6 h-20 flex items-center justify-between">
        <Link href="/" className="flex items-center gap-3 hover:opacity-80 transition-opacity">
          <div className="w-8 h-8 relative flex items-center justify-center">
            <Image src="/openclaw.png" alt="OpenClaw Logo" fill className="object-contain" />
          </div>
          <span className="font-bold text-xl text-white tracking-tight">ClawdFace</span>
        </Link>

        <div className="hidden md:flex items-center gap-10 text-base font-medium text-zinc-300">
          <Link href="/#how-it-works" className="hover:text-[#00E3AA] transition-colors">How it Works</Link>
          <Link href="/#features" className="hover:text-[#00E3AA] transition-colors">Features</Link>
          <Link href="/#pricing" className="hover:text-[#00E3AA] transition-colors">Pricing</Link>
          <Link href="/#demo" className="hover:text-[#00E3AA] transition-colors">Live Demo</Link>
          <Link href="/#contact" className="hover:text-[#00E3AA] transition-colors">Contact Us</Link>
        </div>

        <div className="flex items-center gap-6">
          {user ? (
            <>
              <Link href="/dashboard">
                <motion.button 
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.95 }}
                  className="bg-[#00E3AA] text-black px-7 py-3 rounded-xl text-base font-bold tracking-tight hover:shadow-[0_0_20px_rgba(0,227,170,0.3)] transition-all"
                >
                  Dashboard
                </motion.button>
              </Link>
              <button 
                onClick={async () => {
                  try {
                    await stackApp.signOut();
                    window.location.href = "/";
                  } catch (error) {
                    console.error("Logout failed:", error);
                  }
                }}
                className="text-base font-medium text-zinc-400 hover:text-white transition-colors"
              >
                Log Out
              </button>
            </>
          ) : (
            <>
              <Link href="/handler/sign-in" className="text-base font-medium text-white hover:text-[#00E3AA] transition-colors">
                Log In
              </Link>
              <Link href="/handler/sign-up">
                <motion.button 
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.95 }}
                  className="bg-[#00E3AA] text-black px-7 py-3 rounded-xl text-base font-bold tracking-tight hover:shadow-[0_0_20px_rgba(0,227,170,0.3)] transition-all"
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
