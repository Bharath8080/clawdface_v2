"use client";

import { motion } from "framer-motion";

export function DemoSection() {
  return (
    <section id="demo" className="py-14 md:py-24 px-6 md:px-8 bg-canvas relative overflow-hidden">
      {/* Background dots */}
      {[
        { top: "10%", left: "3%" }, { top: "25%", left: "8%" },
        { top: "60%", left: "4%" }, { top: "80%", left: "10%" },
        { top: "15%", left: "92%" }, { top: "45%", left: "96%" },
        { top: "70%", left: "90%" }, { top: "90%", left: "95%" },
      ].map((dot, i) => (
        <div
          key={i}
          className="absolute w-1 h-1 rounded-full bg-white/15 pointer-events-none"
          style={{ top: dot.top, left: dot.left }}
        />
      ))}

      <div className="max-w-5xl mx-auto relative z-10">
        {/* Title */}
        <motion.h2
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          className="text-4xl md:text-5xl font-black text-white text-center mb-12"
        >
          See Clawd Face In{" "}
          <span className="text-brand">Action</span>
        </motion.h2>

        {/* App window */}
        <motion.div
          initial={{ opacity: 0, y: 24 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ delay: 0.1 }}
          className="rounded-2xl overflow-hidden shadow-2xl"
          style={{ background: "#1c1c1e" }}
        >
          {/* Title bar */}
          <div className="flex items-center justify-between px-5 py-4" style={{ background: "#2c2c2e" }}>
            {/* Traffic lights */}
            <div className="flex items-center gap-2">
              <div className="w-3.5 h-3.5 rounded-full bg-[#FF5F56]" />
              <div className="w-3.5 h-3.5 rounded-full bg-[#FFBD2E]" />
              <div className="w-3.5 h-3.5 rounded-full bg-[#27C93F]" />
            </div>
            {/* Center label */}
            <div className="flex items-center gap-2">
              <span className="w-2 h-2 rounded-full bg-red-500 animate-pulse" />
              <span className="text-zinc-300 text-sm font-medium">Clawd - Your Agent</span>
            </div>
            {/* Timer */}
            <span className="text-zinc-500 text-sm font-mono">00:09:98</span>
          </div>

          {/* Video area */}
          <div className="relative bg-black" style={{ aspectRatio: "16/8" }}>
            {/* Main video (dark/agent area) */}
            <div className="absolute inset-0 bg-surface-secondary" />

            {/* Bottom-left agent label */}
            <div className="absolute bottom-5 left-5 flex items-center gap-2">
              <span className="w-2.5 h-2.5 rounded-full bg-brand animate-pulse" />
              <span className="text-white text-sm font-medium">Clawd - Your Agent</span>
            </div>

            {/* You — PiP bottom-right */}
            <div
              className="absolute bottom-5 right-5 rounded-2xl overflow-hidden w-[120px] md:w-[180px]"
              style={{ background: "#1e1e3a" }}
            >
              <div className="px-3 pt-2.5 pb-1">
                <span className="text-white text-xs font-semibold">You</span>
              </div>
              {/* Avatar placeholder */}
              <div className="flex items-center justify-center pb-4 pt-2 relative">
                <div className="w-16 h-16 rounded-full bg-[#2e2e5e] flex items-center justify-center">
                  <svg width="36" height="36" viewBox="0 0 24 24" fill="#5a5a9a">
                    <path d="M12 12c2.7 0 4.8-2.1 4.8-4.8S14.7 2.4 12 2.4 7.2 4.5 7.2 7.2 9.3 12 12 12zm0 2.4c-3.2 0-9.6 1.6-9.6 4.8v2.4h19.2v-2.4c0-3.2-6.4-4.8-9.6-4.8z"/>
                  </svg>
                </div>
                {/* Orange notification badge */}
                <div className="absolute bottom-3 right-5 w-7 h-7 rounded-full bg-orange-500 flex items-center justify-center shadow-lg">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="white">
                    <path d="M12 2a3 3 0 0 0-3 3v7a3 3 0 0 0 6 0V5a3 3 0 0 0-3-3z"/>
                    <path d="M19 10v2a7 7 0 0 1-14 0v-2" stroke="white" strokeWidth="2" fill="none" strokeLinecap="round"/>
                  </svg>
                </div>
              </div>
            </div>
          </div>

          {/* Control bar */}
          <div
            className="flex items-center justify-center gap-4 py-5"
            style={{ background: "#2c2c2e" }}
          >
            {/* + */}
            <button className="w-12 h-12 rounded-full bg-[#3a3a3c] flex items-center justify-center hover:bg-[#48484a] transition-colors">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="white">
                <line x1="12" y1="5" x2="12" y2="19" stroke="white" strokeWidth="2.5" strokeLinecap="round"/>
                <line x1="5" y1="12" x2="19" y2="12" stroke="white" strokeWidth="2.5" strokeLinecap="round"/>
              </svg>
            </button>

            {/* Share / upload */}
            <button className="w-12 h-12 rounded-full bg-[#3a3a3c] flex items-center justify-center hover:bg-[#48484a] transition-colors">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <polyline points="16 16 12 12 8 16"/>
                <line x1="12" y1="12" x2="12" y2="21"/>
                <path d="M20.39 18.39A5 5 0 0 0 18 9h-1.26A8 8 0 1 0 3 16.3"/>
              </svg>
            </button>

            {/* Mic off */}
            <button className="w-12 h-12 rounded-full bg-[#3a3a3c] flex items-center justify-center hover:bg-[#48484a] transition-colors">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <line x1="1" y1="1" x2="23" y2="23"/>
                <path d="M9 9v3a3 3 0 0 0 5.12 2.12M15 9.34V4a3 3 0 0 0-5.94-.6"/>
                <path d="M17 16.95A7 7 0 0 1 5 12v-2m14 0v2a7 7 0 0 1-.11 1.23"/>
                <line x1="12" y1="19" x2="12" y2="23"/>
                <line x1="8" y1="23" x2="16" y2="23"/>
              </svg>
            </button>

            {/* Camera off */}
            <button className="w-12 h-12 rounded-full bg-[#3a3a3c] flex items-center justify-center hover:bg-[#48484a] transition-colors">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <line x1="1" y1="1" x2="23" y2="23"/>
                <path d="M21 21H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h3m3-3h6l2 3h3a2 2 0 0 1 2 2v9.34"/>
                <circle cx="12" cy="12" r="3" stroke="white" strokeWidth="2" fill="none"/>
              </svg>
            </button>

            {/* End call */}
            <button className="w-12 h-12 rounded-full bg-red-500 flex items-center justify-center hover:bg-red-600 transition-colors shadow-lg shadow-red-500/30">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                <path d="M10.68 13.31a16 16 0 0 0 3.41 2.6l1.27-1.27a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7 2 2 0 0 1 1.72 2v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6 19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L8.09 9.91"/>
              </svg>
            </button>
          </div>
        </motion.div>
      </div>
    </section>
  );
}
