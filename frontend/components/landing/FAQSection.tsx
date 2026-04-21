"use client";

import { useState } from "react";
import { motion, AnimatePresence } from "framer-motion";

const FAQS = [
  {
    question: "Does ClawdFace give your bot a human-like avatar voice?",
    answer:
      "Yes. ClawdFace integrates ElevenLabs for industry-leading neural voice synthesis. Your AI avatar speaks with natural tone, pacing, and emotion — completely indistinguishable from a human voice in most cases.",
  },
  {
    question: "Can I use ClawdFace agent for meetings?",
    answer:
      "Absolutely. ClawdFace avatars can join live video calls on Zoom, Google Meet, Microsoft Teams, and more. Your AI participant shows up with a face, speaks, and interacts just like a real attendee.",
  },
  {
    question: "How does ClawdFace integrate with my existing bot?",
    answer:
      "Zero code changes required. Paste your OpenClaw endpoint URL and API key into ClawdFace settings. We handle the WebRTC signaling, STT/TTS pipeline, and avatar rendering — your OpenClaw logic stays exactly as-is.",
  },
  {
    question: "Is ClawdFace free to start?",
    answer:
      "Yes. Our free tier includes 10 minutes of total usage to test everything end-to-end. No credit card required. Upgrade to a paid plan when you&apos;re ready to go live with more minutes and concurrent sessions.",
  },
  {
    question: "Can ClawdFace agents use tools and return assistant results?",
    answer:
      "Yes. Because ClawdFace connects directly to your OpenClaw agent, it supports full tool use — web search, API calls, database queries, and any custom actions you've configured in your Clawdbot.",
  },
];

function FAQItem({ question, answer }: { question: string; answer: string }) {
  const [open, setOpen] = useState(false);

  return (
    <div className="border-b border-white/5 last:border-0">
      <button
        onClick={() => setOpen(!open)}
        className="w-full flex items-center justify-between py-5 text-left gap-6 group"
      >
        <span className={`text-base font-semibold transition-colors ${open ? "text-[#00E3AA]" : "text-white group-hover:text-[#00E3AA]"}`}>
          {question}
        </span>
        <div
          className={`flex-shrink-0 w-8 h-8 rounded-full border flex items-center justify-center transition-all duration-300 ${
            open
              ? "bg-[#00E3AA]/10 border-[#00E3AA]/30 rotate-45"
              : "bg-white/5 border-white/10"
          }`}
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke={open ? "#00E3AA" : "white"} strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
            <line x1="12" y1="5" x2="12" y2="19" />
            <line x1="5" y1="12" x2="19" y2="12" />
          </svg>
        </div>
      </button>

      <AnimatePresence>
        {open && (
          <motion.div
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: "auto", opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ duration: 0.25 }}
            className="overflow-hidden"
          >
            <p className="text-zinc-400 leading-relaxed pb-5 text-sm">{answer}</p>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}

export function FAQSection() {
  return (
    <section className="py-24 px-6 bg-[#050505]">
      <div className="max-w-3xl mx-auto">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          className="text-center mb-14"
        >
          <div className="inline-flex items-center gap-2 bg-white/5 border border-white/10 px-4 py-2 rounded-full mb-6">
            <span className="text-zinc-400 text-sm font-semibold">Got Questions?</span>
          </div>
          <h2 className="text-4xl md:text-5xl font-bold text-white mb-4">
            Frequently Asked Questions
          </h2>
          <p className="text-xl text-zinc-400">
            Everything you need to know about ClawdFace.
          </p>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ delay: 0.1 }}
          className="rounded-2xl border border-white/10 bg-[#111] px-6 divide-y divide-white/5"
        >
          {FAQS.map((faq, idx) => (
            <FAQItem key={idx} {...faq} />
          ))}
        </motion.div>
      </div>
    </section>
  );
}
