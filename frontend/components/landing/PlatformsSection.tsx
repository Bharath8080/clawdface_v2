"use client";

import Image from "next/image";
import { motion } from "framer-motion";

const PLATFORMS = [
  {
    name: "Zoom",
    icon: "/icons/zoom-icon.svg",
    style: {
      background: "radial-gradient(circle at center, #0c1428 0%, #162050 60%, #1a2a6a 100%)",
      border: "1.5px solid rgba(59,130,246,0.45)",
      boxShadow: "0 0 28px 4px rgba(37,99,235,0.22), inset 0 0 20px rgba(37,99,235,0.08)",
    },
  },
  {
    name: "Google Meet",
    icon: "/icons/google-meet-icon.svg",
    style: {
      background: "radial-gradient(circle at center, #071a0e 0%, #0d3018 60%, #0f4020 100%)",
      border: "1.5px solid rgba(34,197,94,0.5)",
      boxShadow: "0 0 32px 6px rgba(22,163,74,0.28), inset 0 0 20px rgba(22,163,74,0.10)",
    },
  },
  {
    name: "MS Teams",
    icon: "/icons/ms-teams-icon.svg",
    style: {
      background: "radial-gradient(circle at center, #110d2e 0%, #1e1545 60%, #261860 100%)",
      border: "1.5px solid rgba(124,58,237,0.45)",
      boxShadow: "0 0 28px 4px rgba(109,40,217,0.22), inset 0 0 20px rgba(109,40,217,0.08)",
    },
  },
  {
    name: "Website",
    icon: "/icons/website-icon.svg",
    style: {
      background: "radial-gradient(circle at center, #141414 0%, #1e1e1e 60%, #252525 100%)",
      border: "1.5px solid rgba(255,255,255,0.12)",
      boxShadow: "0 0 20px 2px rgba(255,255,255,0.04)",
    },
  },
];

export function PlatformsSection() {
  return (
    <section className="relative py-14 md:py-24 px-6 overflow-hidden">
      {/* Scattered dots */}
      {[
        { top: "10%", left: "2%" }, { top: "30%", left: "5%" },
        { top: "65%", left: "3%" }, { top: "85%", left: "7%" },
        { top: "15%", left: "93%" }, { top: "50%", left: "96%" },
        { top: "78%", left: "94%" },
      ].map((d, i) => (
        <div key={i} className="absolute w-1 h-1 rounded-full bg-white/10 pointer-events-none"
          style={{ top: d.top, left: d.left }} />
      ))}

      <div className="relative z-10 max-w-4xl mx-auto text-center">
        {/* Heading */}
        <motion.h2
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          className="text-[32px] font-semibold leading-8 mb-14"
        >
          <span className="text-white">Talk &amp; Interact to Clawd Face</span>
          <span className="text-white/40"> Anywhere You Meet</span>
        </motion.h2>

        {/* Platform icons */}
        <div className="flex flex-wrap items-center justify-center gap-10 md:gap-16 mb-14">
          {PLATFORMS.map((p, idx) => (
            <motion.div
              key={idx}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ delay: idx * 0.1 }}
              className="flex flex-col items-center gap-4"
            >
              <div
                className="w-[130px] h-[130px] rounded-full flex items-center justify-center transition-transform duration-300 hover:scale-105"
                style={p.style}
              >
                <Image
                  src={p.icon}
                  alt={p.name}
                  width={52}
                  height={52}
                  className="object-contain"
                />
              </div>
              <span className="text-zinc-400 text-sm font-medium tracking-wide">{p.name}</span>
            </motion.div>
          ))}
        </div>

        {/* Description */}
        <motion.p
          initial={{ opacity: 0, y: 12 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ delay: 0.3 }}
          className="text-body text-lg leading-7 max-w-2xl mx-auto"
        >
          You can invite your Open Claw agent anywhere to meetings.&nbsp; Clawd Face into
          the tools you already use. Invite it to your calls and interact in real time — no
          tab-switching, no copy-pasting.
        </motion.p>
      </div>
    </section>
  );
}
