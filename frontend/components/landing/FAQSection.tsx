"use client";

import { useState } from "react";
import { motion, AnimatePresence } from "framer-motion";

const CARD_BG = "#111113";

const FAQS = [
  {
    question: "Does Clawd Face gives both human-like face and voice?",
    answer:
      "Yes! You can give it a natural voice and a friendly face to make interactions feel real.",
  },
  {
    question: "Can my ClawdFace Agent join meetings?",
    answer:
      "Absolutely. ClawdFace avatars can join live video calls on Zoom, Google Meet, Microsoft Teams, and more. Your AI participant shows up with a face, speaks, and interacts just like a real attendee.",
  },
  {
    question: "How does the ClawdFace Agent handle emails?",
    answer:
      "Your ClawdFace Agent gets its own email address. You can invite it to meetings or send it tasks directly, and it will respond and take action just like a human teammate.",
  },
  {
    question: "Is My Data Safe?",
    answer:
      "Yes. ClawdFace is GDPR, HIPAA, ISO 42001, and AICPA SOC 2 compliant. We never allow third-party AI providers to store or train on your data.",
  },
  {
    question: "Can my ClawdFace Agent see and respond visually?",
    answer:
      "Yes. ClawdFace supports vision — your agent can see what's on screen or shared in a video call and respond contextually in real time.",
  },
  {
    question: "What platforms can I use my AI on?",
    answer:
      "ClawdFace works on Zoom, Google Meet, Microsoft Teams, and directly on your website as an embedded avatar. More platform integrations are coming soon.",
  },
];

function FAQItem({
  index,
  question,
  answer,
  defaultOpen = false,
}: {
  index: number;
  question: string;
  answer: string;
  defaultOpen?: boolean;
}) {
  const [open, setOpen] = useState(defaultOpen);

  return (
    <motion.div
      initial={{ opacity: 0, y: 16 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true }}
      transition={{ delay: index * 0.08 }}
      className="rounded-2xl overflow-hidden"
      style={{ background: CARD_BG, border: "1px solid rgba(255,255,255,0.07)" }}
    >
      <button
        onClick={() => setOpen(!open)}
        className="w-full flex items-center justify-between px-5 md:px-7 py-5 text-left gap-4"
      >
        <span className="text-white font-semibold text-base leading-snug">
          {index + 1}. {question}
        </span>
        <div className="shrink-0">
          {open ? (
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <polyline points="18 15 12 9 6 15" />
            </svg>
          ) : (
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="rgba(255,255,255,0.4)" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <polyline points="6 9 12 15 18 9" />
            </svg>
          )}
        </div>
      </button>

      <AnimatePresence initial={false}>
        {open && (
          <motion.div
            key="answer"
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: "auto", opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ duration: 0.22 }}
            className="overflow-hidden"
          >
            <p className="text-zinc-400 text-sm leading-relaxed px-5 md:px-7 pb-6">
              {answer}
            </p>
          </motion.div>
        )}
      </AnimatePresence>
    </motion.div>
  );
}

export function FAQSection() {
  return (
    <section className="py-14 md:py-24 px-6 bg-canvas relative overflow-hidden">
      {/* Ambient glow — right side reddish (matching screenshot) */}
      <div className="pointer-events-none absolute top-1/3 right-0 w-[400px] h-[400px] rounded-full blur-[160px]"
        style={{ background: "rgba(180,60,30,0.12)" }} />

      {/* Scattered dots */}
      {[
        { top: "5%",  left: "2%"  }, { top: "20%", left: "5%"  },
        { top: "55%", left: "3%"  }, { top: "80%", left: "7%"  },
        { top: "10%", left: "93%" }, { top: "35%", left: "96%" },
        { top: "65%", left: "94%" }, { top: "88%", left: "92%" },
      ].map((d, i) => (
        <div key={i} className="absolute w-1 h-1 rounded-full bg-white/10 pointer-events-none"
          style={{ top: d.top, left: d.left }} />
      ))}

      <div className="relative z-10 max-w-3xl mx-auto">
        {/* Header */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          className="text-center mb-12"
        >
          <div className="inline-flex items-center px-4 py-1.5 rounded-md bg-brand-subtle border border-brand-muted/20 mb-6">
            <span className="text-brand-muted text-sm font-mono font-semibold tracking-wide">
              Question &amp; Answers
            </span>
          </div>
          <h2 className="text-3xl md:text-[54px] font-bold text-white mb-5 capitalize leading-tight">
            Frequently Asked Questions
          </h2>
          <p className="text-body text-base md:text-lg max-w-3xl mx-auto leading-7">
            Explore how your AI works, what it can do, and how it fits seamlessly into your workflow—
            everything you need to get started with confidence.
          </p>
        </motion.div>

        {/* FAQ items */}
        <div className="space-y-3">
          {FAQS.map((faq, idx) => (
            <FAQItem
              key={idx}
              index={idx}
              question={faq.question}
              answer={faq.answer}
              defaultOpen={idx === 0}
            />
          ))}
        </div>
      </div>
    </section>
  );
}
