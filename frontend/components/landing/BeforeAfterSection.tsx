"use client";

import Image from "next/image";
import { motion } from "framer-motion";

const CARD_BG = "#111113";

const BEFORE_BULLETS = [
  "Cold, impersonal interactions",
  "Can't build rapport with users",
  "Creates user frustration",
  "No real connection whatsoever",
  "Users drop off quickly",
];

const AFTER_BULLETS = [
  "Warm, personal face-to-face experience",
  "Builds trust and rapport instantly",
  "Users feel heard and understood",
  "Real connection, every conversation",
  "Higher engagement and retention",
];

// ── Before chat mockup ─────────────────────────────────────────────────────
const CHAT_MESSAGES = [
  { text: "Hey, can you summarize this report for me?", user: true },
  { text: "...", user: false },
  { text: "Sorry, I didn’t get that. Could you clarify?", user: false },
  { text: "The sales report from last quarter.", user: true },
  { text: "I’m not sure I understand. Can you reword that?", user: false },
  { text: "Ugh… just give me the key points!", user: true },
];

function BeforeChatVisual() {
  return (
    <div
      className="mx-0 rounded-2xl overflow-hidden mb-5"
      style={{ background: "#18181b", border: "1px solid rgba(255,255,255,0.07)" }}
    >
      <div className="p-4 space-y-2">
        {CHAT_MESSAGES.map((msg, i) => (
          <div key={i} className={`flex items-end gap-2 ${msg.user ? "justify-end" : "justify-start"}`}>
            {!msg.user && (
              <div className="w-6 h-6 rounded-full bg-zinc-700 flex items-center justify-center shrink-0">
                <Image src="/icons/agent-chat.svg" alt="bot" width={12} height={12} className="object-contain" />
              </div>
            )}
            {msg.user ? (
              <div className="px-3 py-2 bg-red-400/20 rounded-tl-2xl rounded-tr-2xl rounded-bl-2xl outline outline-[0.5px] outline-offset-[-0.5px] outline-red-500 max-w-[72%]">
                <span className="text-neutral-200 text-sm leading-6">{msg.text}</span>
              </div>
            ) : (
              <span className="text-neutral-100 text-sm font-normal leading-6 max-w-[72%]">{msg.text}</span>
            )}
            {msg.user && (
              <div className="w-6 h-6 rounded-full bg-zinc-700 flex items-center justify-center shrink-0">
                <Image src="/icons/person-icon.svg" alt="user" width={12} height={12} className="object-contain" />
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}

// ── After browser mockup ───────────────────────────────────────────────────
function AfterBrowserVisual() {
  return (
    <div
      className="mx-0 rounded-2xl overflow-hidden mb-5"
      style={{ background: "#18181b", border: "1px solid rgba(255,255,255,0.07)" }}
    >
      {/* Title bar */}
      <div className="flex items-center gap-1.5 px-3 py-2.5" style={{ background: "#222" }}>
        <div className="w-2.5 h-2.5 rounded-full bg-red-500" />
        <div className="w-2.5 h-2.5 rounded-full bg-yellow-400" />
        <div className="w-2.5 h-2.5 rounded-full bg-green-500" />
      </div>
      {/* Avatar area */}
      <div className="relative" style={{ aspectRatio: "16/9" }}>
        <Image
          src="https://assets.trugen.ai/images/avatarImages/priya-wide.jpg"
          alt="Priya avatar"
          fill
          className="object-cover object-top"
          sizes="500px"
        />
        {/* PiP bottom-right */}
        <div className="absolute bottom-3 right-3 w-12 h-12 rounded-xl overflow-hidden border border-white/20 shadow-lg bg-zinc-900" />
        {/* Control bar */}
        <div
          className="absolute bottom-0 left-0 right-0 flex items-center justify-center gap-3 py-2"
          style={{ background: "rgba(0,0,0,0.55)", backdropFilter: "blur(6px)" }}
        >
          <div className="w-7 h-7 rounded-full bg-white/15 flex items-center justify-center">
            <Image src="/icons/carbon_microphone-off-filled.svg" alt="mic" width={14} height={14} className="object-contain" />
          </div>
          <div className="w-7 h-7 rounded-full bg-red-500/80 flex items-center justify-center">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="white" stroke="none">
              <rect x="2" y="2" width="20" height="20" rx="4" />
            </svg>
          </div>
          <div className="w-7 h-7 rounded-full bg-white/15 flex items-center justify-center">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M15 10l5-5m0 0h-4m4 0v4" />
              <path d="M9 14l-5 5m0 0h4m-4 0v-4" />
            </svg>
          </div>
        </div>
      </div>
    </div>
  );
}

export function BeforeAfterSection() {
  return (
    <section className="py-14 md:py-24 px-6 bg-canvas relative overflow-hidden">
      {/* Scattered dots */}
      {[
        { top: "8%",  left: "2%"  }, { top: "25%", left: "5%"  },
        { top: "62%", left: "3%"  }, { top: "85%", left: "7%"  },
        { top: "10%", left: "93%" }, { top: "40%", left: "96%" },
        { top: "72%", left: "94%" }, { top: "90%", left: "92%" },
      ].map((d, i) => (
        <div key={i} className="absolute w-1 h-1 rounded-full bg-white/10 pointer-events-none"
          style={{ top: d.top, left: d.left }} />
      ))}

      <div className="relative z-10 max-w-5xl mx-auto">
        {/* Heading */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          className="text-center mb-14"
        >
          <h2 className="text-3xl md:text-[56px] font-bold leading-tight capitalize">
            <span className="text-white">Before </span>
            <span className="text-body text-2xl md:text-[42px] font-bold">vs </span>
            <span className="text-brand">After</span>
          </h2>
          <p className="text-body text-base md:text-lg mt-4 max-w-xl mx-auto leading-7">
            See what changes when your bot gets a face.
          </p>
        </motion.div>

        {/* Two cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
          {/* LEFT — Text-Based AI */}
          <motion.div
            initial={{ opacity: 0, x: -24 }}
            whileInView={{ opacity: 1, x: 0 }}
            viewport={{ once: true }}
            className="rounded-3xl p-6 flex flex-col"
            style={{ background: CARD_BG, border: "1px solid rgba(255,255,255,0.07)" }}
          >
            {/* Header */}
            <div className="flex items-center gap-3 mb-5">
              <Image src="/icons/majesticons_chat-line.svg" alt="chat" width={32} height={32} className="object-contain" />
              <span className="text-sm font-bold tracking-widest text-zinc-300 uppercase">Text-Based AI</span>
            </div>

            <BeforeChatVisual />

            {/* Subheading */}
            <p className="text-white font-bold text-lg mb-4">Frustrated. Slow. Disconnected.</p>

            {/* Bullets */}
            <div className="space-y-3">
              {BEFORE_BULLETS.map((item, i) => (
                <div key={i} className="flex items-center gap-3">
                  <Image src="/icons/oui_cross-in-circle-filled.svg" alt="x" width={18} height={18} className="shrink-0 object-contain" />
                  <span className="text-zinc-400 text-sm">{item}</span>
                </div>
              ))}
            </div>
          </motion.div>

          {/* RIGHT — Clawd Face */}
          <motion.div
            initial={{ opacity: 0, x: 24 }}
            whileInView={{ opacity: 1, x: 0 }}
            viewport={{ once: true }}
            className="rounded-3xl p-6 flex flex-col"
            style={{ background: CARD_BG, border: "1px solid rgba(255,255,255,0.07)" }}
          >
            {/* Header */}
            <div className="flex items-center gap-3 mb-5">
              <Image src="/icons/lets-icons_video.svg" alt="video" width={32} height={32} className="object-contain" />
              <span className="text-sm font-bold tracking-widest text-zinc-300 uppercase">Clawd Face</span>
            </div>

            <AfterBrowserVisual />

            {/* Subheading */}
            <p className="text-brand font-bold text-lg mb-4">Natural. Instant. Human.</p>

            {/* Bullets */}
            <div className="space-y-3">
              {AFTER_BULLETS.map((item, i) => (
                <div key={i} className="flex items-center gap-3">
                  <Image src="/icons/healthicons_yes.svg" alt="check" width={18} height={18} className="shrink-0 object-contain" />
                  <span className="text-zinc-300 text-sm">{item}</span>
                </div>
              ))}
            </div>
          </motion.div>
        </div>
      </div>
    </section>
  );
}
