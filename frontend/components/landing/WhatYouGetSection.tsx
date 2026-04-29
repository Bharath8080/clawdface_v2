"use client";

import Image from "next/image";
import { motion } from "framer-motion";

// ── Face visual ────────────────────────────────────────────────────────────
function FaceVisual() {
  return (
    <div className="w-44 h-44 relative mx-auto shrink-0">
      <div
        className="absolute w-32 h-32 left-[24px] top-[16px] rounded-full border-[4px] border-red-400"
        style={{ background: "conic-gradient(from 125deg at 50% 50%, #FF7855 77deg, #FFD9CF 209deg, #FE5634 360deg)" }}
      />
      <div className="absolute w-32 h-32 left-[24px] top-[16px] rounded-full overflow-hidden border-[1.3px] border-red-400 shadow-[0px_-4px_24px_0px_rgba(252,212,11,0.16),0px_4px_24px_0px_rgba(242,114,37,0.16),-4px_0px_24px_0px_rgba(158,10,249,0.16),inset_0px_0px_8px_0px_rgba(255,255,255,1.00)]">
        <Image src="https://assets.trugen.ai/images/avatarImages/chole-wide.jpeg" alt="Face" fill className="object-cover object-top" sizes="128px" />
      </div>
      <div className="absolute w-9 h-9 left-[115px] top-[107px] bg-white rounded-full outline outline-1 outline-red-400 flex items-center justify-center overflow-hidden">
        <Image src="/icons/start-icon.svg" alt="start" width={18} height={18} className="object-contain" />
      </div>
    </div>
  );
}

// ── Voice visual ───────────────────────────────────────────────────────────
function VoiceVisual() {
  return (
    <div className="h-44 md:self-stretch md:flex-1 md:min-h-0 flex flex-col items-center justify-center gap-4 overflow-hidden">
      <div className="relative w-20 h-20 shrink-0">
        <div className="absolute w-28 h-24 -left-[14px] -top-[4px] bg-[conic-gradient(from_180deg,rgba(46,197,147,0.20)_0deg,rgba(255,255,255,0.20)_121deg,rgba(0,172,71,0.20)_250deg,rgba(10,255,173,0.20)_360deg)] rounded-full blur-lg" />
        <div className="absolute inset-0 rounded-full overflow-hidden">
          <div className="w-full h-full bg-zinc-950 relative overflow-hidden">
            <div className="absolute w-24 h-24 left-[30px] top-[12px] bg-emerald-500 rounded-full blur-xl" />
            <div className="absolute w-24 h-24 left-[16px] top-[64px] bg-emerald-400 rounded-full blur-xl" />
            <div className="absolute w-14 h-36 left-[-12px] top-[-56px] bg-emerald-400 rounded-full blur-xl" />
            <div className="absolute w-14 h-36 left-[120px] top-[-34px] bg-indigo-400 rounded-full blur-xl" />
          </div>
        </div>
        <div className="absolute w-14 h-14 left-[15px] top-[15px] bg-[radial-gradient(ellipse_50%_50%_at_50%_50%,rgba(255,255,255,0.50)_0%,rgba(255,255,255,0)_100%)] rounded-full blur-sm" />
      </div>
      <Image src="/icons/waves.svg" alt="waveform" width={180} height={36} className="object-contain opacity-60 shrink-0" />
    </div>
  );
}

// ── Vision visual ─────────────────────────────────────────────────────────
function VisionVisual() {
  return (
    <div className="h-44 md:self-stretch md:flex-1 md:min-h-0 flex items-end justify-center overflow-hidden">
      <Image
        src="/icons/call-preview-2.png"
        alt="call preview"
        width={320}
        height={200}
        className="w-full h-auto object-contain object-bottom"
      />
    </div>
  );
}

// ── Email visual ──────────────────────────────────────────────────────────
function EmailVisual() {
  return (
    <div className="h-44 md:self-stretch md:flex-1 md:min-h-0 relative rounded-xl overflow-hidden">
      <div className="absolute inset-0 bg-neutral-900 rounded-xl" />
      <div className="absolute top-0 left-0 right-0 px-2.5 py-2.5 bg-stone-900 flex items-center gap-3">
        <div className="flex gap-[3px]">
          <div className="w-1.5 h-1.5 bg-red-400 rounded-full" />
          <div className="w-1.5 h-1.5 bg-amber-400 rounded-full" />
          <div className="w-1.5 h-1.5 bg-green-500 rounded-full" />
        </div>
        <div className="flex flex-1 gap-[3px]">
          <div className="w-24 h-1.5 bg-stone-900 rounded-full" />
          <div className="flex-1 h-1.5 bg-stone-900 rounded-full" />
        </div>
      </div>
      {[26, 62, 98].map((top, i) => (
        <div key={i} className="absolute left-0 right-0 px-3 py-2 flex justify-between items-center" style={{ top }}>
          <div className="flex items-center gap-3">
            <div className="w-6 h-6 bg-neutral-800 rounded-full border border-neutral-700 shrink-0" />
            <div className="flex flex-col gap-1.5">
              <div className="h-1.5 bg-neutral-700 rounded-3xl" style={{ width: i === 0 ? 80 : i === 1 ? 64 : 96 }} />
              <div className="h-1.5 w-14 bg-neutral-800 rounded-3xl" />
            </div>
          </div>
          <div className="h-1.5 w-10 bg-neutral-800 rounded-3xl" />
        </div>
      ))}
    </div>
  );
}

// ── Meetings visual ───────────────────────────────────────────────────────
function MeetingsVisual() {
  return (
    <div className="h-44 md:self-stretch md:flex-1 md:min-h-0 relative rounded-xl overflow-hidden">
      {/* Main window */}
      <div className="absolute inset-0 bg-neutral-900 rounded-xl" />
      {/* Title bar */}
      <div className="absolute top-0 left-0 right-0 px-3 py-2.5 bg-stone-900 flex items-center gap-4">
        <div className="flex gap-1">
          <div className="w-2 h-2 bg-red-400 rounded-full" />
          <div className="w-2 h-2 bg-amber-400 rounded-full" />
          <div className="w-2 h-2 bg-green-500 rounded-full" />
        </div>
        <div className="flex flex-1 gap-1">
          <div className="w-20 h-1.5 bg-stone-900 rounded-full" />
          <div className="flex-1 h-1.5 bg-stone-900 rounded-full" />
        </div>
      </div>
      {/* Video area */}
      <div className="absolute left-3 right-3 top-10 bottom-8 bg-stone-900 rounded-lg flex gap-2 overflow-hidden">
        <div className="flex-1 bg-neutral-900 rounded-l-lg" />
        <div className="w-16 bg-neutral-900 rounded-r-lg" />
      </div>
      {/* Controls */}
      <div className="absolute bottom-2 left-1/2 -translate-x-1/2 flex items-center gap-1.5">
        <div className="w-4 h-4 bg-white/5 rounded-full outline outline-[0.22px] outline-white/80 flex items-center justify-center">
          <div className="w-2 h-2 border border-white rounded-sm" />
        </div>
        <div className="w-4 h-4 bg-white/5 rounded-full outline outline-[0.22px] outline-white/80 flex items-center justify-center">
          <div className="w-2 h-1.5 bg-white rounded-sm" />
        </div>
        <div className="w-4 h-4 bg-white rounded-full flex items-center justify-center">
          <div className="w-1.5 h-1.5 bg-black rounded-full" />
        </div>
        <div className="w-4 h-4 bg-white rounded-full flex items-center justify-center">
          <div className="w-1.5 h-1.5 bg-black rounded-full" />
        </div>
        <div className="w-4 h-4 bg-red-600 rounded-full flex items-center justify-center">
          <div className="w-2 h-0.5 bg-white rounded" />
        </div>
      </div>
    </div>
  );
}

// ── Card label ────────────────────────────────────────────────────────────
function CardLabel({ name, desc }: { name: string; desc: string }) {
  return (
    <p className="text-sm font-semibold leading-snug shrink-0">
      <span className="text-white">{name} </span>
      <span className="text-white/40">{desc}</span>
    </p>
  );
}

const cardClass = "bg-gradient-to-br from-white/5 to-black rounded-[32px] outline outline-[1.5px] outline-offset-[-1.5px] overflow-hidden cursor-default";
const cardStyle = { outlineColor: "rgba(255,255,255,0)", transition: "outline-color 0.3s ease, box-shadow 0.3s ease" };
const onCardEnter = (e: React.MouseEvent<HTMLDivElement>) => {
  const el = e.currentTarget as HTMLDivElement;
  el.style.outlineColor = "rgba(255,255,255,0.18)";
  el.style.boxShadow = "0 8px 40px rgba(0,0,0,0.5), 0 0 24px rgba(255,255,255,0.04)";
};
const onCardLeave = (e: React.MouseEvent<HTMLDivElement>) => {
  const el = e.currentTarget as HTMLDivElement;
  el.style.outlineColor = "rgba(255,255,255,0)";
  el.style.boxShadow = "none";
};

export function WhatYouGetSection() {
  return (
    <section className="py-14 md:py-24 px-6 relative overflow-hidden">

      {[
        { top: "8%", left: "2%" }, { top: "25%", left: "5%" },
        { top: "62%", left: "3%" }, { top: "85%", left: "7%" },
        { top: "10%", left: "93%" }, { top: "40%", left: "96%" },
        { top: "72%", left: "94%" }, { top: "90%", left: "92%" },
      ].map((d, i) => (
        <div key={i} className="absolute w-1 h-1 rounded-full bg-white/10 pointer-events-none"
          style={{ top: d.top, left: d.left }} />
      ))}

      <div className="relative z-10 max-w-5xl mx-auto flex flex-col items-center gap-14">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          className="flex flex-col items-center gap-5 text-center"
        >
          <h2 className="text-3xl md:text-[56px] font-bold text-white capitalize">What You Get</h2>
          <p className="text-body text-base md:text-lg max-w-xl leading-7">
            Your AI isn&apos;t just smart - it&apos;s a teammate you can interact with anytime, anywhere.
          </p>
        </motion.div>

        <div className="w-full flex flex-col gap-4">
          {/* Top row */}
          <div className="w-full flex flex-col md:flex-row gap-4 md:h-72">
            <motion.div initial={{ opacity: 0, y: 20 }} whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }} transition={{ delay: 0 }} whileHover={{ y: -5 }}
              style={cardStyle} onMouseEnter={onCardEnter} onMouseLeave={onCardLeave}
              className={`flex-1 min-w-0 px-5 pt-6 pb-4 flex flex-col items-center gap-3 ${cardClass}`}>
              <CardLabel name="Face" desc="– Give your AI a human-like presence." />
              <FaceVisual />
            </motion.div>

            <motion.div initial={{ opacity: 0, y: 20 }} whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }} transition={{ delay: 0.1 }} whileHover={{ y: -5 }}
              style={cardStyle} onMouseEnter={onCardEnter} onMouseLeave={onCardLeave}
              className={`flex-1 min-w-0 px-5 py-6 flex flex-col gap-3 ${cardClass}`}>
              <CardLabel name="Voice" desc="– Speak naturally and be understood." />
              <VoiceVisual />
            </motion.div>

            <motion.div initial={{ opacity: 0, y: 20 }} whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }} transition={{ delay: 0.2 }} whileHover={{ y: -5 }}
              style={cardStyle} onMouseEnter={onCardEnter} onMouseLeave={onCardLeave}
              className={`flex-1 min-w-0 px-5 py-6 flex flex-col gap-3 ${cardClass}`}>
              <CardLabel name="Vision" desc="– See and respond visually like a real Assistant." />
              <VisionVisual />
            </motion.div>
          </div>

          {/* Bottom row */}
          <div className="w-full flex flex-col md:flex-row gap-4 md:h-64">
            <motion.div initial={{ opacity: 0, y: 20 }} whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }} transition={{ delay: 0.3 }} whileHover={{ y: -5 }}
              style={cardStyle} onMouseEnter={onCardEnter} onMouseLeave={onCardLeave}
              className={`flex-1 px-5 py-6 flex flex-col gap-3 ${cardClass}`}>
              <CardLabel name="Email" desc="– Your Agent will get an email, which you can use it to invite for Meetings." />
              <EmailVisual />
            </motion.div>

            <motion.div initial={{ opacity: 0, y: 20 }} whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }} transition={{ delay: 0.4 }} whileHover={{ y: -5 }}
              style={cardStyle} onMouseEnter={onCardEnter} onMouseLeave={onCardLeave}
              className={`flex-1 px-5 py-6 flex flex-col gap-3 ${cardClass}`}>
              <CardLabel name="Join Meetings" desc="– Agent can talk to you in Zoom, Google Meet, and more." />
              <MeetingsVisual />
            </motion.div>
          </div>
        </div>
      </div>
    </section>
  );
}
