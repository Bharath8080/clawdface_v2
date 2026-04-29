"use client";

import Image from "next/image";
import { motion } from "framer-motion";

const FEATURES = [
  {
    icon: "/icons/feature-avatar-icon.svg",
    title: "Realistic Human-to-AI Interactions",
    description:
      "Experience hyper-realistic, human-like conversations with AI avatars that respond with natural expressions and fluid movement",
  },
  {
    icon: "/icons/feature-voice-icon.svg",
    title: "Precision Voice & Speed",
    description:
      "Engineered with Deepgram for sub-300ms Speech-to-Text and ElevenLabs for industry-leading, high-fidelity neural voices.",
  },
  {
    icon: "/icons/feature-chat-icon.svg",
    title: "Beyond Text & Audio",
    description:
      "Move past traditional text chatbots and voice-only assistants; engage with your AI through a life-like, immersive video presence.",
  },
  {
    icon: "/icons/feature-monitor-icon.svg",
    title: "Immersive Web Interaction",
    description:
      "A state-of-the-art web platform built specifically for seamless, face-to-face avatar interactions without any complex setup.",
  },
];

export function FeaturesSection() {
  return (
    <section id="features" className="px-6 md:px-20 py-14 md:py-24 relative overflow-hidden scroll-mt-24">

      {/* Scattered dots */}
      {[
        { top: "8%",  left: "8%"  }, { top: "22%", left: "12%" },
        { top: "55%", left: "6%"  }, { top: "78%", left: "14%" },
        { top: "10%", left: "88%" }, { top: "38%", left: "92%" },
        { top: "68%", left: "90%" }, { top: "88%", left: "86%" },
      ].map((d, i) => (
        <div key={i} className="absolute w-0.5 h-0.5 rounded-full bg-zinc-600 pointer-events-none"
          style={{ top: d.top, left: d.left }} />
      ))}

      <div className="relative z-10 max-w-[1104px] mx-auto flex flex-col items-center gap-10 md:gap-16">
        {/* Heading */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          className="flex flex-col items-center gap-6 text-center"
        >
          <h2 className="text-3xl md:text-6xl font-bold max-w-[784px]">
            <span className="text-white">Same OpenClaw. Entirely </span>
            <span className="text-[#FF5C5C]">New Experience</span>
            <span className="text-white">.</span>
          </h2>
          <p className="text-neutral-400 text-base md:text-lg font-normal max-w-[674px] leading-7">
            ClawdFace bridges your custom LLM providers directly to LiveKit-powered
            interactive avatars.
          </p>
        </motion.div>

        {/* 2×2 grid */}
        <div className="self-stretch flex flex-col gap-6">
          {[FEATURES.slice(0, 2), FEATURES.slice(2, 4)].map((row, rowIdx) => (
            <div key={rowIdx} className="flex flex-col md:flex-row gap-6">
              {row.map((f, idx) => (
                <motion.div
                  key={idx}
                  initial={{ opacity: 0, y: 24 }}
                  whileInView={{ opacity: 1, y: 0 }}
                  viewport={{ once: true }}
                  transition={{ delay: (rowIdx * 2 + idx) * 0.1 }}
                  whileHover={{ y: -5 }}
                  className="flex-1 px-6 pt-6 pb-10 bg-gradient-to-br from-white/5 to-black rounded-2xl outline outline-[1.5px] outline-offset-[-1.5px] flex flex-col gap-3.5 cursor-default"
                  style={{
                    outlineColor: "rgba(255,255,255,0.08)",
                    boxShadow: "inset 0px -16.5px 36.9px 0px rgba(255,255,255,0.04)",
                    transition: "outline-color 0.3s ease, box-shadow 0.3s ease",
                  }}
                  onMouseEnter={e => {
                    const el = e.currentTarget as HTMLDivElement;
                    el.style.outlineColor = "rgba(255,255,255,0.22)";
                    el.style.boxShadow = "inset 0px -16.5px 36.9px 0px rgba(255,255,255,0.04), 0 0 28px rgba(255,255,255,0.05), 0 8px 32px rgba(0,0,0,0.5)";
                  }}
                  onMouseLeave={e => {
                    const el = e.currentTarget as HTMLDivElement;
                    el.style.outlineColor = "rgba(255,255,255,0.08)";
                    el.style.boxShadow = "inset 0px -16.5px 36.9px 0px rgba(255,255,255,0.04)";
                  }}
                >
                  <div className="flex flex-col gap-5">
                    {/* Icon pill */}
                    <div
                     
                    >
                      <Image src={f.icon} alt={f.title} width={50} height={50} className="object-contain" />
                    </div>
                    <h3 className="text-white text-xl md:text-2xl font-bold capitalize">{f.title}</h3>
                    <p className="text-zinc-400 text-base font-normal leading-6">{f.description}</p>
                  </div>
                </motion.div>
              ))}
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
