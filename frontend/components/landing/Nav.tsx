"use client";

import Link from "next/link";
import Image from "next/image";
import { motion } from "framer-motion";
import { useUser, useStackApp } from "@stackframe/stack";
import { useEffect, useState } from "react";

export function Nav() {
  const user = useUser();
  const stackApp = useStackApp();
  const [scrolled, setScrolled] = useState(false);

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 24);
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  return (
    <motion.nav
      className="sticky top-0 inset-x-0 z-50 transition-all duration-300"
      animate={{
        backgroundColor: scrolled ? "rgba(6,13,9,0.80)" : "rgba(6,13,9,0)",
        borderBottomColor: scrolled ? "rgba(255,255,255,0.07)" : "rgba(255,255,255,0)",
        backdropFilter: scrolled ? "blur(16px)" : "blur(0px)",
        boxShadow: scrolled ? "0 4px 32px rgba(0,0,0,0.35)" : "none",
      }}
      style={{ borderBottomWidth: 1, borderBottomStyle: "solid" }}
    >
      <div className="max-w-7xl mx-auto px-6 h-[68px] flex items-center justify-between">
        {/* Logo */}
        <Link href="/" className="flex items-center gap-2 hover:opacity-90 transition-opacity">
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img src="/clawdface-logo.svg" alt="ClawdFace" className="w-14 h-12 object-contain shrink-0" />
          <Image src="/icons/ClawdFace-main.svg" alt="ClawdFace" width={100} height={16} className="object-contain" />
        </Link>

        {/* Nav links */}
        <div className="hidden md:flex items-center gap-8 text-[14px] font-medium text-zinc-400">
          <Link href="/#how-it-works" className="hover:text-white transition-colors">How It Works?</Link>
          <Link href="/#features" className="hover:text-white transition-colors">Features</Link>
          <Link href="/#pricing" className="hover:text-white transition-colors">Pricing</Link>
          <Link href="/#contact" className="hover:text-white transition-colors">Contact Us</Link>
        </div>

        {/* Auth */}
        <div className="flex items-center gap-4">
          {user ? (
            <>
              <Link href="/dashboard">
                <motion.button
                  whileHover={{ scale: 1.03 }}
                  whileTap={{ scale: 0.97 }}
                  className="bg-brand text-black px-5 py-2 rounded-2xl text-[14px] font-bold"
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
                className="text-[14px] font-medium text-zinc-400 hover:text-white transition-colors"
              >
                Log Out
              </button>
            </>
          ) : (
            <>
              <Link href="/log-in" className="text-[14px] font-medium text-zinc-400 hover:text-white transition-colors">
                Log In
              </Link>
              <Link href="/sign-up">
                <motion.button
                  whileHover={{ scale: 1.03 }}
                  whileTap={{ scale: 0.97 }}
                  className="bg-brand text-black px-5 py-2 rounded-lg text-[14px] font-bold"
                >
                  Sign Up
                </motion.button>
              </Link>
            </>
          )}
        </div>
      </div>
    </motion.nav>
  );
}
