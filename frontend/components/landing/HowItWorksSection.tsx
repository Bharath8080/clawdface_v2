"use client";

import { motion } from "framer-motion";
import Image from "next/image";

const STEPS = [
  {
    number: "01",
    title: "Connect Your Clawd Agent",
    description:
      "Paste your OpenClaw endpoint URL and API key. ClawdFace instantly connects to your existing bot — no rewrites, no refactoring.",
  },
  {
    number: "02",
    title: "Add Face, Voice & Vision",
    description:
      "Choose an avatar, configure the voice with ElevenLabs, and enable camera vision. Your bot gains a real face in under 2 minutes.",
  },
  {
    number: "03",
    title: "It's Ready to Talk",
    description:
      "Share the session link. Users join a face-to-face video call with your AI — it sees them, hears them, and responds naturally.",
  },
];

export function HowItWorksSection() {
  return (
    <section id="how-it-works" className="py-24 px-6 bg-canvas scroll-mt-24">
      <div className="max-w-7xl mx-auto">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          className="text-center mb-16"
        >
          <div className="inline-flex items-center gap-2 bg-white/5 border border-white/10 px-4 py-2 rounded-full mb-6">
            <span className="text-zinc-400 text-sm font-semibold">3 Simple Steps</span>
          </div>
          <h2 className="text-4xl md:text-5xl font-bold text-white mb-4">How It Works</h2>
          <p className="text-xl text-zinc-400 max-w-2xl mx-auto">
            From your existing OpenClaw bot to a face-to-face AI experience in minutes.
          </p>
        </motion.div>

        <div className="flex flex-col lg:flex-row items-center gap-16">
          {/* Steps */}
          <div className="w-full lg:w-1/2 space-y-8">
            {STEPS.map((step, idx) => (
              <motion.div
                key={idx}
                initial={{ opacity: 0, x: -20 }}
                whileInView={{ opacity: 1, x: 0 }}
                viewport={{ once: true }}
                transition={{ delay: idx * 0.15 }}
                className="flex items-start gap-6 group"
              >
                <div className="flex-shrink-0 w-14 h-14 rounded-2xl bg-brand/10 border border-brand/20 flex items-center justify-center group-hover:bg-brand/20 transition-colors">
                  <span className="text-brand font-bold text-lg font-mono">{step.number}</span>
                </div>
                <div className="flex-1">
                  <h3 className="text-xl font-bold text-white mb-2">{step.title}</h3>
                  <p className="text-zinc-400 leading-relaxed">{step.description}</p>
                </div>
              </motion.div>
            ))}
          </div>

          {/* Avatar preview */}
          <motion.div
            initial={{ opacity: 0, x: 20 }}
            whileInView={{ opacity: 1, x: 0 }}
            viewport={{ once: true }}
            transition={{ delay: 0.2 }}
            className="w-full lg:w-1/2 flex justify-center"
          >
            <div className="w-full max-w-md rounded-3xl overflow-hidden border border-white/10 bg-surface-card shadow-2xl relative">
              <div className="relative aspect-[3/4]">
                <Image
                  src="https://assets.trugen.ai/images/avatarImages/chole-wide.jpeg"
                  alt="ClawdFace avatar interaction"
                  fill
                  className="object-cover object-top"
                  sizes="480px"
                />
                <div className="absolute inset-0 bg-gradient-to-t from-surface-card via-transparent to-transparent" />
              </div>
              {/* Status overlay */}
              <div className="absolute bottom-0 inset-x-0 p-6">
                <div className="bg-black/60 backdrop-blur-md rounded-2xl p-4 border border-white/10">
                  <div className="flex items-center gap-3 mb-3">
                    <div className="w-3 h-3 rounded-full bg-brand animate-pulse" />
                    <span className="text-brand text-sm font-semibold">Ready to Talk</span>
                  </div>
                  <p className="text-white text-sm font-medium mb-1">Chloe · AI Agent</p>
                  <p className="text-zinc-400 text-xs">OpenClaw Connected · ElevenLabs Voice · Vision Enabled</p>
                </div>
              </div>
            </div>
          </motion.div>
        </div>
      </div>
    </section>
  );
}
