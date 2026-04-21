"use client";

import { motion } from "framer-motion";

const PLATFORMS = [
  {
    name: "OpenClaw",
    color: "bg-[#00E3AA]/20 border-[#00E3AA]/30",
    icon: (
      <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#00E3AA" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M12 2L2 7l10 5 10-5-10-5z" />
        <path d="M2 17l10 5 10-5" />
        <path d="M2 12l10 5 10-5" />
      </svg>
    ),
  },
  {
    name: "LiveKit",
    color: "bg-blue-500/20 border-blue-500/30",
    icon: (
      <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#60a5fa" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <polygon points="23 7 16 12 23 17 23 7" />
        <rect x="1" y="5" width="15" height="14" rx="2" ry="2" />
      </svg>
    ),
  },
  {
    name: "ElevenLabs",
    color: "bg-purple-500/20 border-purple-500/30",
    icon: (
      <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#c084fc" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M12 2a3 3 0 0 0-3 3v7a3 3 0 0 0 6 0V5a3 3 0 0 0-3-3z" />
        <path d="M19 10v2a7 7 0 0 1-14 0v-2" />
        <line x1="12" y1="19" x2="12" y2="23" />
        <line x1="8" y1="23" x2="16" y2="23" />
      </svg>
    ),
  },
  {
    name: "Deepgram",
    color: "bg-orange-500/20 border-orange-500/30",
    icon: (
      <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#fb923c" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="12" cy="12" r="2" />
        <path d="M16.24 7.76a6 6 0 0 1 0 8.49m-8.48-.01a6 6 0 0 1 0-8.49m11.31-2.82a10 10 0 0 1 0 14.14m-14.14 0a10 10 0 0 1 0-14.14" />
      </svg>
    ),
  },
  {
    name: "Web SDK",
    color: "bg-pink-500/20 border-pink-500/30",
    icon: (
      <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#f472b6" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <polyline points="16 18 22 12 16 6" />
        <polyline points="8 6 2 12 8 18" />
      </svg>
    ),
  },
  {
    name: "REST API",
    color: "bg-yellow-500/20 border-yellow-500/30",
    icon: (
      <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#fbbf24" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M18 20V10" />
        <path d="M12 20V4" />
        <path d="M6 20v-6" />
      </svg>
    ),
  },
];

export function PlatformsSection() {
  return (
    <section className="py-20 px-6 bg-[#0a0a0a] overflow-hidden">
      <div className="max-w-7xl mx-auto">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          className="text-center mb-12"
        >
          <h2 className="text-3xl md:text-4xl font-bold text-white mb-4">
            Talk & Interact to Clawd Face Anywhere You Meet
          </h2>
          <p className="text-zinc-400 max-w-xl mx-auto">
            Built on best-in-class infrastructure so your AI runs everywhere with zero latency.
          </p>
        </motion.div>

        <div className="flex flex-wrap items-center justify-center gap-6 md:gap-10">
          {PLATFORMS.map((platform, idx) => (
            <motion.div
              key={idx}
              initial={{ opacity: 0, scale: 0.8 }}
              whileInView={{ opacity: 1, scale: 1 }}
              viewport={{ once: true }}
              transition={{ delay: idx * 0.08 }}
              className="flex flex-col items-center gap-3"
            >
              <div
                className={`w-20 h-20 rounded-full border-2 flex items-center justify-center ${platform.color} transition-all duration-300 hover:scale-110`}
              >
                {platform.icon}
              </div>
              <span className="text-zinc-400 text-sm font-medium">{platform.name}</span>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}
