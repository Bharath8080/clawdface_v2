"use client";

import Image from "next/image";
import { motion } from "framer-motion";

const CARD_BG = "#0c0e1e";
const CARD_BORDER = "rgba(255,255,255,0.08)";

// ── Card 1 visual: plug icon in dark square with dot-grid bg ──────────────
function ConnectVisual() {
  return (
    <div className="relative flex items-center justify-center h-[200px]">
      {/* Dot grid */}
      <svg className="absolute inset-0 w-full h-full opacity-20" xmlns="http://www.w3.org/2000/svg">
        <defs>
          <pattern id="dotgrid" width="28" height="28" patternUnits="userSpaceOnUse">
            <circle cx="1" cy="1" r="1" fill="rgba(255,255,255,0.35)" />
          </pattern>
        </defs>
        <rect width="100%" height="100%" fill="url(#dotgrid)" />
      </svg>
      {/* Plug icon box */}
      <div className="relative z-10 w-[88px] h-[88px] rounded-2xl bg-[#111320] border border-white/10 flex items-center justify-center shadow-xl">
        <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
          <path d="M12 22v-5" />
          <path d="M9 8V2" />
          <path d="M15 8V2" />
          <path d="M18 8H6a2 2 0 0 0-2 2v3a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2v-3a2 2 0 0 0-2-2z" />
        </svg>
      </div>
    </div>
  );
}

// ── Card 2 visual: silhouette + waveform bars ─────────────────────────────
const WAVE_L = [18, 28, 22, 36, 26, 40, 24, 32, 20];
const WAVE_R = [20, 32, 24, 40, 26, 36, 22, 28, 18];

function VoiceVisual() {
  return (
    <div className="relative flex items-center justify-center gap-4 h-[200px]">
      {/* Left bars */}
      <div className="flex items-center gap-[3px]">
        {WAVE_L.map((h, i) => (
          <div key={i} className="w-[3px] rounded-full bg-white/20" style={{ height: h }} />
        ))}
      </div>
      {/* Head silhouette */}
      <div className="w-[90px] h-[90px] rounded-full bg-white/10 border border-white/15 flex items-center justify-center">
        <svg width="46" height="46" viewBox="0 0 24 24" fill="rgba(255,255,255,0.35)" stroke="none">
          <path d="M12 12c2.7 0 4.8-2.1 4.8-4.8S14.7 2.4 12 2.4 7.2 4.5 7.2 7.2 9.3 12 12 12zm0 2.4c-3.2 0-9.6 1.6-9.6 4.8v2.4h19.2v-2.4c0-3.2-6.4-4.8-9.6-4.8z" />
        </svg>
      </div>
      {/* Right bars */}
      <div className="flex items-center gap-[3px]">
        {WAVE_R.map((h, i) => (
          <div key={i} className="w-[3px] rounded-full bg-white/20" style={{ height: h }} />
        ))}
      </div>
      {/* Cursor */}
      <div className="absolute bottom-7 right-10">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="white">
          <path d="M4 4l7.07 17 2.51-7.39L21 11.07z" />
        </svg>
      </div>
    </div>
  );
}

// ── Card 3 visual: circular avatar + floating badges ─────────────────────
function ReadyVisual() {
  return (
    <div className="relative flex items-center justify-center h-[200px]">
      {/* Circular avatar */}
      <div className="relative w-[130px] h-[130px] rounded-full overflow-hidden border-2 border-white/15 shadow-xl">
        <Image
          src="https://assets.trugen.ai/images/avatarImages/chole-wide.jpeg"
          alt="Chloe avatar"
          fill
          className="object-cover object-top"
          sizes="130px"
        />
      </div>
      {/* Chat icon — top-left */}
      <div className="absolute top-4 left-8 w-9 h-9 rounded-full bg-[#1a1c2e] border border-white/15 flex items-center justify-center shadow-lg">
        <Image src="/icons/chat.svg" alt="chat" width={16} height={16} className="object-contain" />
      </div>
      {/* Mic icon — top-right */}
      <div className="absolute top-10 right-8 w-9 h-9 rounded-full bg-[#1a1c2e] border border-white/15 flex items-center justify-center shadow-lg">
        <Image src="/icons/microphone.svg" alt="microphone" width={16} height={16} className="object-contain" />
      </div>
      {/* Green send button — bottom-right */}
      <div className="absolute bottom-6 right-12 w-9 h-9 rounded-full bg-brand flex items-center justify-center shadow-lg shadow-brand/30">
        <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="black" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
          <line x1="12" y1="19" x2="12" y2="5" />
          <polyline points="5 12 12 5 19 12" />
        </svg>
      </div>
    </div>
  );
}

// ── Cards config ──────────────────────────────────────────────────────────
const CARDS = [
  {
    visual: <ConnectVisual />,
    title: "Connect Your Clawd Agent",
    description: "Plug in your agent's logic, tools, and workflows - bring your AI to life in minutes.",
  },
  {
    visual: <VoiceVisual />,
    title: "Add Face, Voice & Vision",
    description: "Give your agent a human-like presence: speak naturally, respond visually, and interact like a real teammate.",
  },
  {
    visual: <ReadyVisual />,
    title: "It's Ready To Talk",
    description: "Connect and chat seamlessly - on the website, Zoom, Google Meet, or your favorite platform.",
  },
];

export function HowItWorksSection() {
  return (
    <section id="how-it-works" className="py-14 md:py-24 px-6 scroll-mt-24 relative overflow-hidden">
      {/* Scattered dots */}
      {[
        { top: "8%",  left: "2%"  }, { top: "22%", left: "6%"  },
        { top: "60%", left: "3%"  }, { top: "82%", left: "8%"  },
        { top: "12%", left: "93%" }, { top: "45%", left: "96%" },
        { top: "75%", left: "94%" },
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
          className="text-center mb-14"
        >
          <h2 className="text-3xl md:text-[56px] font-bold text-white mb-5 capitalize">
            How It Works
          </h2>
          <p className="text-body text-base md:text-lg max-w-xl mx-auto leading-7">
            Bring ClawdFace into the tools you already use. Invite it to your calls and
            interact in real time — no tab-switching, no copy-pasting.
          </p>
        </motion.div>

        {/* 3-column cards */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
          {CARDS.map((card, idx) => (
            <motion.div
              key={idx}
              initial={{ opacity: 0, y: 24 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ delay: idx * 0.12 }}
              className="flex flex-col rounded-3xl overflow-hidden"
              style={{ background: CARD_BG, border: `1px solid ${CARD_BORDER}` }}
            >
              <div className="px-6 pt-8">{card.visual}</div>
              <div className="px-7 pb-8 pt-4">
                <h3 className="text-xl font-bold text-white mb-2 leading-snug">{card.title}</h3>
                <p className="text-zinc-500 text-sm leading-relaxed">{card.description}</p>
              </div>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}
