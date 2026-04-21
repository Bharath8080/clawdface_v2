"use client";

import Image from "next/image";
import { motion } from "framer-motion";

const BEFORE_ITEMS = [
  "Frustrated. Slow. Disconnected.",
  "Cold, impersonal interactions",
  "Can't build rapport with users",
  "Creates user frustration",
  "No real connection whatsoever",
];

const AFTER_ITEMS = [
  "Natural. Instant. Human.",
  "Warm, personal face-to-face experience",
  "Builds trust and rapport instantly",
  "Users feel heard and understood",
  "Real connection, every conversation",
];

export function BeforeAfterSection() {
  return (
    <section className="py-24 px-6 bg-[#0a0a0a]">
      <div className="max-w-7xl mx-auto">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          className="text-center mb-16"
        >
          <h2 className="text-4xl md:text-5xl font-bold text-white mb-4">
            Before <span className="text-zinc-500">vs</span> After
          </h2>
          <p className="text-xl text-zinc-400 max-w-xl mx-auto">
            See what changes when your bot gets a face.
          </p>
        </motion.div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Before */}
          <motion.div
            initial={{ opacity: 0, x: -20 }}
            whileInView={{ opacity: 1, x: 0 }}
            viewport={{ once: true }}
            className="rounded-2xl border border-red-500/20 bg-[#111] overflow-hidden"
          >
            <div className="px-6 py-4 border-b border-red-500/10 bg-red-500/5 flex items-center gap-3">
              <div className="w-8 h-8 rounded-full bg-red-500/20 flex items-center justify-center">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#ef4444" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                  <line x1="18" y1="6" x2="6" y2="18" />
                  <line x1="6" y1="6" x2="18" y2="18" />
                </svg>
              </div>
              <span className="text-red-400 font-bold text-sm uppercase tracking-widest">Text-Based AI</span>
            </div>

            {/* Chat mockup */}
            <div className="p-4 border-b border-white/5 bg-[#0d0d0d]">
              <div className="space-y-2">
                {["How do I reset my password?", "Please hold while I look that up...", "Sorry, I couldn't find that.", "Can you try again?"].map(
                  (msg, i) => (
                    <div key={i} className={`flex ${i % 2 === 0 ? "justify-end" : "justify-start"}`}>
                      <div
                        className={`px-3 py-2 rounded-xl text-xs max-w-[75%] ${
                          i % 2 === 0 ? "bg-white/10 text-zinc-300" : "bg-zinc-800 text-zinc-400"
                        }`}
                      >
                        {msg}
                      </div>
                    </div>
                  )
                )}
              </div>
            </div>

            <div className="p-6 space-y-3">
              {BEFORE_ITEMS.map((item, i) => (
                <div key={i} className="flex items-start gap-3">
                  <div className="w-5 h-5 rounded-full bg-red-500/20 flex items-center justify-center flex-shrink-0 mt-0.5">
                    <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="#ef4444" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round">
                      <line x1="18" y1="6" x2="6" y2="18" />
                      <line x1="6" y1="6" x2="18" y2="18" />
                    </svg>
                  </div>
                  <span className={`text-sm ${i === 0 ? "text-red-400 font-semibold" : "text-zinc-500"}`}>
                    {item}
                  </span>
                </div>
              ))}
            </div>
          </motion.div>

          {/* After */}
          <motion.div
            initial={{ opacity: 0, x: 20 }}
            whileInView={{ opacity: 1, x: 0 }}
            viewport={{ once: true }}
            className="rounded-2xl border border-[#00E3AA]/20 bg-[#111] overflow-hidden"
          >
            <div className="px-6 py-4 border-b border-[#00E3AA]/10 bg-[#00E3AA]/5 flex items-center gap-3">
              <div className="w-8 h-8 rounded-full bg-[#00E3AA]/20 flex items-center justify-center">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#00E3AA" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                  <polyline points="20 6 9 17 4 12" />
                </svg>
              </div>
              <span className="text-[#00E3AA] font-bold text-sm uppercase tracking-widest">Clawd Face</span>
            </div>

            {/* Avatar preview */}
            <div className="relative h-48 border-b border-white/5">
              <Image
                src="https://assets.trugen.ai/images/avatarImages/priya-wide.jpg"
                alt="ClawdFace avatar"
                fill
                className="object-cover object-top"
                sizes="600px"
              />
              <div className="absolute inset-0 bg-gradient-to-t from-[#111]/80 via-transparent to-transparent" />
              <div className="absolute bottom-4 left-4 flex items-center gap-2 bg-black/60 backdrop-blur-sm px-3 py-1.5 rounded-full border border-[#00E3AA]/20">
                <span className="w-2 h-2 rounded-full bg-[#00E3AA] animate-pulse" />
                <span className="text-[#00E3AA] text-xs font-semibold">Live · Face-to-Face</span>
              </div>
            </div>

            <div className="p-6 space-y-3">
              {AFTER_ITEMS.map((item, i) => (
                <div key={i} className="flex items-start gap-3">
                  <div className="w-5 h-5 rounded-full bg-[#00E3AA]/20 flex items-center justify-center flex-shrink-0 mt-0.5">
                    <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="#00E3AA" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round">
                      <polyline points="20 6 9 17 4 12" />
                    </svg>
                  </div>
                  <span className={`text-sm ${i === 0 ? "text-[#00E3AA] font-semibold" : "text-zinc-300"}`}>
                    {item}
                  </span>
                </div>
              ))}
            </div>
          </motion.div>
        </div>
      </div>
    </section>
  );
}
