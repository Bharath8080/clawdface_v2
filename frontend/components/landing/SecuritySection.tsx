"use client";

import Image from "next/image";
import { motion } from "framer-motion";


const BADGES = [
  {
    logo: "/icons/gdpr-light 1.svg",
    logoSize: { w: 110, h: 110 },
    title: "GDPR Compliant",
    description:
      "ClawdFace complies with GDPR, protecting your privacy and ensuring that your personal data is always secure and well-managed.",
  },
  {
    logo: "/icons/security2.svg",
    logoSize: { w: 110, h: 110 },
    title: "ISO Compliant",
    description:
      "ClawdFace is ISO 42001 certified, setting a trusted standard for secure, transparent, and responsible AI management in the workplace.",
  },
  {
    logo: "/icons/HIPAA COMPLIANT 3.svg",
    logoSize: { w: 110, h: 110 },
    title: "HIPAA Compliant",
    description:
      "ClawdFace complies with HIPAA, safeguarding sensitive health information and keeping all data confidential, private, and fully secure.",
  },
  {
    logo: "/icons/soc-2-light 1.svg",
    logoSize: { w: 110, h: 110 },
    title: "AICPA SOC 2 Compliant",
    description:
      "ClawdFace complies with AICPA SOC 2, maintaining rigorous controls to protect your data and ensuring AI providers cannot access or use it.",
  },
];

const FEATURES = [
  {
    icon: "/icons/lucide_folder-lock.svg",
    title: "Zero third party data retention",
    description: "We don't allow third-party AI providers to store any of your data.",
  },
  {
    icon: "/icons/iconamoon_shield-yes.svg",
    title: "No third party data training",
    description: "We forbid third party AI providers from training your data.",
  },
  {
    icon: "/icons/octicon_ai-model-24.svg",
    title: "Multi model support",
    description: "We offer the latest AI models, with unified permissions, privacy, and security controls.",
  },
];

// ── Shield visual ──────────────────────────────────────────────────────────
function ShieldVisual() {
  return (
    <div className="relative w-full h-72 md:h-96 overflow-hidden mb-8">
      {/* Frosted glass card — centered */}
      <div className="absolute left-1/2 top-[36px] -translate-x-1/2 w-64 md:w-80 h-72 md:h-96 overflow-hidden">
        <div className="absolute inset-0  " />
        <Image src="/icons/security-shield.svg" alt="shield" width={320} height={384} className="object-contain relative z-10" />
        <div className="absolute inset-0 flex items-center justify-center z-20">
          <Image src="/icons/Union.svg" alt="dots" width={60} height={60} className="object-contain" />
        </div>
        {/* Glowing dots inside card */}
        <div className="absolute w-[5px] h-[5px] left-[256px] top-[236px] bg-zinc-300 rounded-full shadow-[0px_0px_4px_0px_rgba(255,255,255,0.55)]" />
        <div className="absolute w-[5px] h-[5px] left-[294px] top-[147px] bg-zinc-300 rounded-full shadow-[0px_0px_4px_0px_rgba(255,255,255,0.55)]" />
        <div className="absolute w-[5px] h-[5px] left-[213px] top-[64px] bg-zinc-300 rounded-full shadow-[0px_0px_4px_0px_rgba(255,255,255,0.55)]" />
        <div className="absolute w-[5px] h-[5px] left-[224px] top-0 bg-zinc-300 rounded-full shadow-[0px_0px_4px_0px_rgba(255,255,255,0.55)]" />
        <div className="absolute w-0.5 h-0.5 left-[104px] top-[297px] bg-zinc-300 rounded-full shadow-[0px_0px_4px_0px_rgba(255,255,255,0.55)]" />
        <div className="absolute w-0.5 h-0.5 left-[125px] top-[162px] bg-zinc-300 rounded-full shadow-[0px_0px_4px_0px_rgba(255,255,255,0.55)]" />
        <div className="absolute w-1 h-1 left-[27px] top-[186px] bg-zinc-300 rounded-full shadow-[0px_0px_4px_0px_rgba(255,255,255,0.55)]" />
        <div className="absolute w-[3px] h-[3px] left-[37px] top-[171px] bg-zinc-300 rounded-full shadow-[0px_0px_4px_0px_rgba(255,255,255,0.55)]" />
        <div className="absolute w-1 h-1 left-[259px] top-[64px] bg-zinc-300 rounded-full shadow-[0px_0px_4px_0px_rgba(255,255,255,0.55)]" />
        <div className="absolute w-1.5 h-1.5 left-[34px] top-[85px] bg-zinc-300 rounded-full shadow-[0px_0px_4px_0px_rgba(255,255,255,0.55)]" />
        <div className="absolute w-0.5 h-0.5 left-[262px] top-[187px] bg-zinc-300 rounded-full shadow-[0px_0px_4px_0px_rgba(255,255,255,0.55)]" />
      </div>

      {/* Outer scattered dots */}
      <div className="absolute w-1 h-1 right-[22%] top-[133px] bg-zinc-300 rounded-full shadow-[0px_0px_4px_0px_rgba(255,255,255,0.55)]" />
      <div className="absolute w-1 h-1 left-[22%] top-[42px] bg-zinc-300 rounded-full shadow-[0px_0px_4px_0px_rgba(255,255,255,0.55)]" />
      <div className="absolute w-1.5 h-1.5 left-[21%] top-[244px] bg-zinc-300 rounded-full shadow-[0px_0px_4px_0px_rgba(255,255,255,0.55)]" />

      {/* Brand glow — over shield + union icon */}
      <div className="absolute left-1/2 top-[72%] -translate-x-1/2 -translate-y-1/2 w-64 h-64 rounded-full pointer-events-none z-30"
        style={{ background: "radial-gradient(ellipse, rgba(0,227,170,0.28) 0%, rgba(0,227,170,0.10) 5%, transparent 72%)" }} />

      {/* Emerald glow — bottom right */}
      <div className="absolute right-[18%] bottom-0 w-96 h-14 bg-emerald-500/50 blur-[94px]" />

      {/* Bottom fade to canvas */}
      <div
        className="absolute bottom-0 left-0 w-full h-56 pointer-events-none"
        style={{ background: "linear-gradient(to bottom, transparent, #060d09)" }}
      />
    </div>
  );
}

export function SecuritySection() {
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

        {/* Shield visual */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
        >
          <ShieldVisual />
        </motion.div>

        {/* Header */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          className="text-center flex flex-col items-center gap-3 mb-14 -mt-20 md:-mt-40 z-10 relative"
        >
          <p className="text-teal-500 text-base md:text-xl font-semibold capitalize">
            Privacy and Security
          </p>
          <div className="flex flex-col items-center gap-3">
            <h2 className="text-3xl md:text-5xl font-semibold text-white capitalize">
              Enterprise Grade Security
            </h2>
            <p className="text-zinc-500 text-base md:text-lg font-normal max-w-[599px] text-center leading-7">
              Don&apos;t just take our word for it - (TruGen AI) is certified to protect
              sensitive data, no matter the industry or use case.
            </p>
          </div>
        </motion.div>

        {/* 2×2 compliance cards */}
        <div className="flex flex-col gap-6 mb-6">
          {[BADGES.slice(0, 2), BADGES.slice(2, 4)].map((row, rowIdx) => (
            <div key={rowIdx} className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {row.map((badge, idx) => (
                <motion.div
                  key={idx}
                  initial={{ opacity: 0, y: 20 }}
                  whileInView={{ opacity: 1, y: 0 }}
                  viewport={{ once: true }}
                  transition={{ delay: (rowIdx * 2 + idx) * 0.1 }}
                  whileHover={{ y: -5 }}
                  className="flex items-center gap-4 md:gap-6 h-auto md:h-44 p-5 bg-neutral-900 rounded-2xl cursor-default"
                  style={{
                    outline: "1px solid rgba(255,255,255,0.08)",
                    transition: "outline-color 0.3s ease, box-shadow 0.3s ease",
                  }}
                  onMouseEnter={e => {
                    const el = e.currentTarget as HTMLDivElement;
                    el.style.outlineColor = "rgba(255,255,255,0.20)";
                    el.style.boxShadow = "0 8px 36px rgba(0,0,0,0.5), 0 0 20px rgba(255,255,255,0.04)";
                  }}
                  onMouseLeave={e => {
                    const el = e.currentTarget as HTMLDivElement;
                    el.style.outlineColor = "rgba(255,255,255,0.08)";
                    el.style.boxShadow = "none";
                  }}
                >
                  <div className="shrink-0 flex items-center justify-center w-20 h-24 md:w-28 md:h-32">
                    <Image
                      src={badge.logo}
                      alt={badge.title}
                      width={badge.logoSize.w}
                      height={badge.logoSize.h}
                      className="object-contain"
                    />
                  </div>
                  <div className="flex-1 flex flex-col gap-3 md:gap-4">
                    <h3 className="text-white text-xl md:text-2xl font-semibold capitalize leading-snug">{badge.title}</h3>
                    <p className="text-neutral-400 text-sm md:text-base font-normal leading-6">{badge.description}</p>
                  </div>
                </motion.div>
              ))}
            </div>
          ))}
        </div>

        {/* Bottom 3-feature row */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ delay: 0.3 }}
          className="grid grid-cols-1 md:grid-cols-3 rounded-3xl overflow-hidden"
          // style={{ background: CARD_BG, border: "1px solid rgba(255,255,255,0.07)" }}
        >
          {FEATURES.map((f, idx) => (
            <div
              key={idx}
              className={`flex items-start gap-4 p-5 md:p-7 transition-colors duration-200 hover:bg-white/[0.03] cursor-default ${idx < FEATURES.length - 1 ? "border-b md:border-b-0 md:border-r border-white/10" : ""}`}
            >
              <div className="shrink-0 mt-0.5">
                <Image src={f.icon} alt={f.title} width={22} height={22} className="object-contain" />
              </div>
              <div>
                <p className="text-white text-sm font-semibold mb-1 leading-snug">{f.title}</p>
                <p className="text-zinc-500 text-xs leading-relaxed">{f.description}</p>
              </div>
            </div>
          ))}
        </motion.div>
      </div>
    </section>
  );
}
