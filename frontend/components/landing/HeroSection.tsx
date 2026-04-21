"use client";

import { useState, useEffect } from "react";
import { motion, AnimatePresence } from "framer-motion";
import Link from "next/link";
import Image from "next/image";
import { useUser } from "@stackframe/stack";

const AVATARS = [
  { url: "https://assets.trugen.ai/images/avatarImages/chole-wide.jpeg", name: "Chloe" },
  { url: "https://assets.trugen.ai/images/avatarImages/priya-wide.jpg", name: "Priya" },
  { url: "https://assets.trugen.ai/images/avatarImages/aman-wide.jpg", name: "Aman" },
  { url: "https://assets.trugen.ai/images/avatarImages/matt.jpeg", name: "Matt" },
  { url: "https://assets.trugen.ai/images/avatarImages/sameer-wide.jpeg", name: "Sameer" },
];

// Scattered background dots
const DOTS = [
  { top: "15%", left: "5%" }, { top: "28%", left: "12%" }, { top: "8%", left: "22%" },
  { top: "45%", left: "6%" }, { top: "62%", left: "15%" }, { top: "80%", left: "8%" },
  { top: "18%", left: "38%" }, { top: "72%", left: "32%" }, { top: "90%", left: "45%" },
  { top: "12%", left: "55%" }, { top: "85%", left: "60%" }, { top: "35%", left: "70%" },
];

export function HeroSection() {
  const user = useUser();
  const [currentIdx, setCurrentIdx] = useState(0);
  const ROTATION_TIME = 3500;

  useEffect(() => {
    const timer = setInterval(() => {
      setCurrentIdx((prev) => (prev + 1) % AVATARS.length);
    }, ROTATION_TIME);
    return () => clearInterval(timer);
  }, []);

  return (
    <section className="relative min-h-[88vh] flex items-center px-8 py-20 overflow-hidden bg-[#0d0d0d]">
      {/* Subtle background dots */}
      {DOTS.map((dot, i) => (
        <div
          key={i}
          className="absolute w-1 h-1 rounded-full bg-white/20 pointer-events-none"
          style={{ top: dot.top, left: dot.left }}
        />
      ))}
      {/* Faint green ambient glow bottom-left */}
      <div className="absolute bottom-0 left-0 w-[500px] h-[500px] bg-[#00E3AA]/8 rounded-full blur-[150px] pointer-events-none" />

      <div className="max-w-[1400px] mx-auto w-full flex flex-col lg:flex-row items-center gap-12 lg:gap-16 relative z-10">
        {/* ── Left column ── */}
        <div className="w-full lg:w-[42%] flex flex-col items-start">
          {/* Badge */}
          <motion.div
            initial={{ opacity: 0, y: 16 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.45 }}
            className="flex items-center gap-3 mb-8"
          >
            <span className="bg-[#1a2d24] text-[#82e8b2] px-3 py-1 rounded text-sm font-mono font-semibold border border-[#82e8b2]/20">
              OpenClaw
            </span>
            <span className="text-zinc-500 text-sm">→</span>
            <span className="text-zinc-400 text-sm font-mono tracking-wide">Interactive Avatars</span>
          </motion.div>

          {/* Headline */}
          <motion.h1
            initial={{ opacity: 0, y: 16 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.45, delay: 0.1 }}
            className="text-[clamp(3.5rem,7vw,5.5rem)] font-black tracking-tight leading-[0.95] uppercase mb-8"
          >
            <span className="text-white block">Give Your</span>
            <span className="text-white block">Open Claw</span>
            <span className="text-[#FF4747] block">A Face</span>
          </motion.h1>

          {/* Description */}
          <motion.p
            initial={{ opacity: 0, y: 16 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.45, delay: 0.2 }}
            className="text-[15px] text-zinc-400 mb-10 max-w-[480px] leading-[1.7]"
          >
            Your bot handles text. We handle the face and voice. Install the skill, verify your API
            key, and call your Clawdbot like a video call. It sees you, hears you speak, reads the
            transcript, and replies through a lifelike avatar. You change nothing in your OpenClaw
            logic.
          </motion.p>

          {/* CTAs */}
          <motion.div
            initial={{ opacity: 0, y: 16 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.45, delay: 0.3 }}
            className="flex items-center gap-4"
          >
            <Link href={user ? "/dashboard" : "/handler/sign-up"}>
              <button className="px-7 py-3.5 bg-[#00E3AA] text-black rounded-lg font-bold text-[15px] hover:bg-[#00cfA0] transition-colors">
                {user ? "Dashboard" : "Get Started"}
              </button>
            </Link>
            <Link href="#how-it-works">
              <button className="flex items-center gap-2.5 px-6 py-3.5 bg-[#1a1a1a] text-white border border-white/10 rounded-lg font-semibold text-[15px] hover:bg-[#222] transition-colors">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="white">
                  <path d="M8 5v14l11-7z" />
                </svg>
                Watch It Work
              </button>
            </Link>
          </motion.div>
        </div>

        {/* ── Right column: Video card ── */}
        <motion.div
          initial={{ opacity: 0, x: 24 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ duration: 0.6, delay: 0.15 }}
          className="w-full lg:w-[58%]"
        >
          <div className="rounded-2xl border border-white/10 bg-[#1a1a1a] overflow-hidden shadow-2xl">
            {/* Card header */}
            <div className="flex items-center justify-between px-5 py-4 border-b border-white/5">
              <span className="text-xs font-mono text-zinc-500 uppercase tracking-[0.15em]">
                Real Time Interaction
              </span>
              <div className="flex items-center gap-2">
                <div className="w-3 h-3 rounded-full bg-red-500/80" />
                <div className="w-3 h-3 rounded-full bg-yellow-400/80" />
                <div className="w-3 h-3 rounded-full bg-green-500/80" />
              </div>
            </div>

            {/* Avatar image */}
            <div className="relative" style={{ aspectRatio: "16/9" }}>
              <AnimatePresence initial={false}>
                <motion.div
                  key={currentIdx}
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  exit={{ opacity: 0 }}
                  transition={{ duration: 0.5 }}
                  className="absolute inset-0"
                >
                  <Image
                    src={AVATARS[currentIdx].url}
                    alt={AVATARS[currentIdx].name}
                    fill
                    className="object-cover object-top"
                    sizes="(max-width: 1024px) 100vw, 800px"
                    priority
                  />
                </motion.div>
              </AnimatePresence>
            </div>

            {/* Avatar name footer */}
            <div className="px-5 py-4 border-t border-white/5 text-center">
              <span className="text-[#00E3AA] font-semibold text-base">
                {AVATARS[currentIdx].name}
              </span>
              <span className="text-zinc-400 text-base"> (Open Claw Agent)</span>
            </div>
          </div>
        </motion.div>
      </div>
    </section>
  );
}
