"use client";

import Image from "next/image";
import { motion } from "framer-motion";

export function WhatYouGetSection() {
  return (
    <section className="py-24 px-6 bg-surface-secondary">
      <div className="max-w-7xl mx-auto">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          className="text-center mb-16"
        >
          <div className="inline-flex items-center gap-2 bg-white/5 border border-white/10 px-4 py-2 rounded-full mb-6">
            <span className="text-zinc-400 text-sm font-semibold">Full Stack AI Presence</span>
          </div>
          <h2 className="text-4xl md:text-5xl font-bold text-white mb-4">What You Get</h2>
          <p className="text-xl text-zinc-400 max-w-2xl mx-auto">
            Everything your AI needs to show up like a human — face, voice, vision, and more.
          </p>
        </motion.div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {/* Face card — tall, spans full height on left */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            className="md:row-span-2 rounded-2xl border border-white/10 bg-surface-card overflow-hidden group hover:border-brand/30 transition-all duration-300 relative"
          >
            <div className="relative h-72 md:h-80">
              <Image
                src="https://assets.trugen.ai/images/avatarImages/priya-wide.jpg"
                alt="Face feature"
                fill
                className="object-cover object-top group-hover:scale-105 transition-transform duration-500"
                sizes="400px"
              />
              <div className="absolute inset-0 bg-gradient-to-t from-surface-card via-surface-card/30 to-transparent" />
            </div>
            <div className="p-6">
              <div className="w-10 h-10 rounded-xl bg-brand/10 border border-brand/20 flex items-center justify-center mb-4">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#00E3AA" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <circle cx="12" cy="8" r="4" />
                  <path d="M6 20v-2a4 4 0 0 1 4-4h4a4 4 0 0 1 4 4v2" />
                </svg>
              </div>
              <h3 className="text-2xl font-bold text-white mb-2">Face</h3>
              <p className="text-zinc-400 text-sm leading-relaxed">
                Hyper-realistic AI avatars with natural expressions, head movement, and lip-sync — indistinguishable from a real person on a video call.
              </p>
            </div>
          </motion.div>

          {/* Voice card */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ delay: 0.1 }}
            className="rounded-2xl border border-white/10 bg-surface-card p-6 group hover:border-purple-500/30 transition-all duration-300"
          >
            <div className="w-10 h-10 rounded-xl bg-purple-500/10 border border-purple-500/20 flex items-center justify-center mb-4">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#c084fc" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M12 2a3 3 0 0 0-3 3v7a3 3 0 0 0 6 0V5a3 3 0 0 0-3-3z" />
                <path d="M19 10v2a7 7 0 0 1-14 0v-2" />
                <line x1="12" y1="19" x2="12" y2="23" />
              </svg>
            </div>
            <h3 className="text-xl font-bold text-white mb-2">Voice</h3>
            <p className="text-zinc-400 text-sm leading-relaxed mb-4">
              Sub-300ms speech synthesis powered by ElevenLabs — natural tone, emotion, and cadence that feels human.
            </p>
            {/* Waveform visual */}
            <div className="flex items-end gap-1 h-10">
              {[...Array(24)].map((_, i) => (
                <div
                  key={i}
                  className="flex-1 bg-purple-500/40 rounded-full animate-pulse"
                  style={{
                    height: `${20 + ((i * 11) % 80)}%`,
                    animationDelay: `${i * 0.07}s`,
                  }}
                />
              ))}
            </div>
          </motion.div>

          {/* Vision card */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ delay: 0.2 }}
            className="rounded-2xl border border-white/10 bg-surface-card p-6 group hover:border-blue-500/30 transition-all duration-300"
          >
            <div className="w-10 h-10 rounded-xl bg-blue-500/10 border border-blue-500/20 flex items-center justify-center mb-4">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#60a5fa" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
                <circle cx="12" cy="12" r="3" />
              </svg>
            </div>
            <h3 className="text-xl font-bold text-white mb-2">Vision</h3>
            <p className="text-zinc-400 text-sm leading-relaxed">
              Your AI sees the user through their camera — reads context, detects emotion, and responds to what&apos;s actually happening in the room.
            </p>
          </motion.div>

          {/* Email card */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ delay: 0.3 }}
            className="rounded-2xl border border-white/10 bg-surface-card p-6 group hover:border-yellow-500/30 transition-all duration-300"
          >
            <div className="w-10 h-10 rounded-xl bg-yellow-500/10 border border-yellow-500/20 flex items-center justify-center mb-4">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#fbbf24" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z" />
                <polyline points="22,6 12,13 2,6" />
              </svg>
            </div>
            <h3 className="text-xl font-bold text-white mb-2">Email</h3>
            <p className="text-zinc-400 text-sm leading-relaxed">
              Your AI agent can draft, send, and manage emails on behalf of users — fully integrated with your OpenClaw logic.
            </p>
          </motion.div>

          {/* Join Meetings card */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ delay: 0.4 }}
            className="rounded-2xl border border-white/10 bg-surface-card p-6 group hover:border-brand/30 transition-all duration-300"
          >
            <div className="w-10 h-10 rounded-xl bg-brand/10 border border-brand/20 flex items-center justify-center mb-4">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#00E3AA" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <rect x="3" y="4" width="18" height="18" rx="2" ry="2" />
                <line x1="16" y1="2" x2="16" y2="6" />
                <line x1="8" y1="2" x2="8" y2="6" />
                <line x1="3" y1="10" x2="21" y2="10" />
              </svg>
            </div>
            <h3 className="text-xl font-bold text-white mb-2">Join Meetings</h3>
            <p className="text-zinc-400 text-sm leading-relaxed">
              Add your AI avatar to live Zoom, Teams, or Meet calls. It participates, answers questions, and represents your brand — live.
            </p>
          </motion.div>
        </div>
      </div>
    </section>
  );
}
