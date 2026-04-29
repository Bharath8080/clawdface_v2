"use client";

import Image from "next/image";
import { motion } from "framer-motion";

const CARDS = [
  {
    icon: "/icons/Icon.svg",
    title: "Bots Feel Distant",
    description:
      "Chatbots and voice assistants often lack presence, making interactions feel cold and impersonal.",
    highlight: false,
  },
  {
    icon: "/icons/Icon (1).svg",
    title: "Human Feels Personal",
    description:
      "People naturally connect with human-like interaction—it builds comfort, attention, and trust.",
    highlight: true,
  },
  {
    icon: "/icons/Icon (2).svg",
    title: "Make Your Agent Feel Real",
    description:
      "Add a face to your AI to create trust, capture attention, and drive deeper engagement because people respond to presence, not prompts.",
    highlight: false,
  },
];

export function WhyFaceMattersSection() {
  return (
    <section className="relative py-14 md:py-24 px-6 overflow-hidden">

      {/* Scattered dots */}
      {[
        { top: "8%", left: "2%" }, { top: "20%", left: "6%" },
        { top: "55%", left: "3%" }, { top: "75%", left: "8%" },
        { top: "12%", left: "94%" }, { top: "40%", left: "97%" },
        { top: "68%", left: "92%" }, { top: "88%", left: "96%" },
      ].map((d, i) => (
        <div key={i} className="absolute w-1 h-1 rounded-full bg-white/10 pointer-events-none"
          style={{ top: d.top, left: d.left }} />
      ))}

      <div className="relative z-10 max-w-6xl mx-auto">
        {/* Header */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          className="text-center mb-16"
        >
          <div className="inline-flex items-center px-4 py-1.5 rounded-md bg-brand-subtle border border-brand-muted/20 mb-6">
            <span className="text-brand-muted text-sm font-mono font-semibold tracking-wide">
              What Matters Most
            </span>
          </div>
          <h2 className="text-5xl md:text-6xl font-black text-white mb-5 tracking-tight font-outfit">
            Why Face Matters
          </h2>
          <p className="text-zinc-400 text-lg max-w-2xl mx-auto leading-relaxed">
            Bring Clawd Face into the tools you already use. Invite it to your calls and
            interact in real time — no tab-switching, no copy-pasting.
          </p>
        </motion.div>

        {/* Cards */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
          {CARDS.map((card, idx) => (
            <motion.div
              key={idx}
              initial={{ opacity: 0, y: 24 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ delay: idx * 0.1 }}
              whileHover={{ y: -6 }}
              className="relative flex flex-col p-6 md:p-8 rounded-3xl overflow-hidden cursor-default"
              style={{
                background: "var(--color-surface-card, #111113)",
                border: card.highlight
                  ? "1px solid rgba(0,227,170,0.20)"
                  : "1px solid rgba(255,255,255,0.08)",
                transition: "border-color 0.3s ease, box-shadow 0.3s ease",
              }}
              onMouseEnter={e => {
                const el = e.currentTarget as HTMLDivElement;
                if (card.highlight) {
                  el.style.borderColor = "rgba(0,227,170,0.50)";
                  el.style.boxShadow = "0 8px 36px rgba(0,0,0,0.45), 0 0 32px rgba(0,227,170,0.12)";
                } else {
                  el.style.borderColor = "rgba(255,255,255,0.20)";
                  el.style.boxShadow = "0 8px 36px rgba(0,0,0,0.45), 0 0 20px rgba(255,255,255,0.04)";
                }
              }}
              onMouseLeave={e => {
                const el = e.currentTarget as HTMLDivElement;
                el.style.borderColor = card.highlight
                  ? "rgba(0,227,170,0.20)"
                  : "rgba(255,255,255,0.08)";
                el.style.boxShadow = "none";
              }}
            >
              {/* Teal bottom glow for highlighted card */}
              {card.highlight && (
                <div className="pointer-events-none absolute bottom-0 left-0 right-0 h-40 bg-gradient-to-t from-brand/20 via-brand/5 to-transparent rounded-b-3xl" />
              )}

              {/* Icon circle */}
              <div className="w-[72px] h-[72px] rounded-full flex items-center justify-center mb-8 shrink-0">
                <Image
                  src={card.icon}
                  alt={card.title}
                  width={70}
                  height={70}
                  className="object-contain"
                />
              </div>

              {/* Text */}
              <h3 className="text-xl font-bold text-white mb-3 leading-snug">
                {card.title}
              </h3>
              <p className="text-zinc-400 text-sm leading-relaxed">
                {card.description}
              </p>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}
