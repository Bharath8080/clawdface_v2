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

// ... imports remain the same

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
    <section className="relative min-h-[90vh] flex items-center px-6 py-24 overflow-hidden bg-[#0a0a0a]">
      {/* Scattered background dots */}
      {DOTS.map((dot, i) => (
        <div
          key={i}
          className="absolute w-1 h-1 rounded-full bg-white/20 pointer-events-none"
          style={{ top: dot.top, left: dot.left }}
        />
      ))}
      
      {/* Green ambient glow */}
      <div className="absolute top-[10%] left-[-10%] w-[700px] h-[700px] rounded-full pointer-events-none"
        style={{ background: "radial-gradient(circle, rgba(0,227,170,0.15) 0%, rgba(0,227,170,0.05) 50%, transparent 75%)" }} />
      <div className="absolute top-[30%] left-[5%] w-[400px] h-[400px] rounded-full pointer-events-none"
        style={{ background: "radial-gradient(circle, rgba(0,227,170,0.1) 0%, transparent 70%)" }} />
      {/* Orange accent glow */}
      <div className="absolute top-[40%] left-[20%] w-[280px] h-[280px] rounded-full pointer-events-none"
        style={{ background: "radial-gradient(circle, rgba(255,92,92,0.18) 0%, rgba(255,92,92,0.06) 50%, transparent 75%)" }} />

      <div className="max-w-7xl mx-auto w-full flex flex-col lg:flex-row items-center gap-12 lg:gap-16 relative z-10">
        {/* ── Left column ── */}
        <div className="w-full lg:w-[48%] flex flex-col items-start">
          {/* Badge */}
          <motion.div
            initial={{ opacity: 0, y: 16 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.45 }}
            className="flex items-center gap-3 mb-10"
          >
            <span className="bg-[#1a2f2a] text-[#00e3aa] px-3 py-1.5 rounded-md text-xs font-mono font-medium border border-[#00e3aa]/20">
              OpenClaw
            </span>
            <span className="text-zinc-600 text-sm">→</span>
            <span className="text-zinc-500 text-sm font-mono tracking-wide">Interactive Avatars</span>
          </motion.div>

          {/* Headline - ALL CAPS */}
          <motion.h1
            initial={{ opacity: 0, y: 16 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.45, delay: 0.1 }}
            className="text-[clamp(3rem,6vw,5rem)] font-black tracking-tight leading-[0.9] mb-8"
          >
            <span className="text-white block uppercase">Give Your</span>
            <span className="text-white block uppercase">Open Claw</span>
            {/* Coral/Salmon color for "A FACE" */}
            <span className="block uppercase" style={{ color: "#ff6b6b" }}>A Face</span>
          </motion.h1>

          {/* Description */}
          <motion.p
            initial={{ opacity: 0, y: 16 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.45, delay: 0.2 }}
            className="text-[15px] text-zinc-400 mb-10 max-w-[500px] leading-[1.7]"
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
            className="flex items-center flex-wrap gap-4"
          >
            <Link href={user ? "/dashboard" : "/handler/sign-up"}>
              <button className="px-7 py-3.5 bg-[#00e3aa] text-black rounded-lg font-bold text-[15px] hover:bg-[#00c995] transition-colors">
                {user ? "Dashboard" : "Get Started"}
              </button>
            </Link>
            <Link href="#how-it-works">
              <button className="flex items-center gap-2.5 px-6 py-3.5 bg-[#1a1a1a] text-white border border-white/10 rounded-lg font-semibold text-[15px] hover:bg-[#222] transition-colors">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor">
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
          className="w-full lg:w-[52%]"
        >
          <div className="p-6 bg-neutral-900 rounded-2xl outline outline-1 outline-offset-[-1px] outline-zinc-800 flex flex-col gap-6 shadow-2xl">
            {/* Card header */}
            <div className="flex items-center justify-between">
              <span className="text-zinc-400 text-sm md:text-base font-semibold font-mono uppercase tracking-wide">
                Real Time Interaction
              </span>
              <div className="flex items-center gap-1.5">
                <div className="w-2.5 h-2.5 rounded-full bg-red-400" />
                <div className="w-2.5 h-2.5 rounded-full bg-amber-400" />
                <div className="w-2.5 h-2.5 rounded-full bg-green-500" />
              </div>
            </div>

            {/* Avatar image */}
            <div className="relative w-full h-56 sm:h-72 md:h-96 rounded-lg overflow-hidden">
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
                    sizes="(max-width: 1024px) 100vw, 600px"
                    priority
                  />
                </motion.div>
              </AnimatePresence>
            </div>

            {/* Avatar name footer */}
            <div>
              <span className="text-teal-500 text-lg font-semibold">
                {AVATARS[currentIdx].name}
              </span>
              <span className="text-neutral-400 text-lg font-semibold"> (Open Claw Agent)</span>
            </div>
          </div>
        </motion.div>
      </div>
    </section>
  );
}
