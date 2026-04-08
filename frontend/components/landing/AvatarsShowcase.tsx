"use client";

import { motion } from "framer-motion";
import Image from "next/image";
import { useState } from "react";

const avatars = [
  {
    id: "lisa",
    name: "Lisa",
    role: "Support Bot",
    image: "/avatars/lisa.png",
  },
  {
    id: "chloe",
    name: "Chloe",
    role: "Career Coach",
    image: "/avatars/chole.jpeg",
  },
  {
    id: "aman",
    name: "Aman",
    role: "AI Interviewer",
    image: "/avatars/aman.jpg",
  },
  {
    id: "matt",
    name: "Matt",
    role: "Assistant",
    image: "/avatars/kevin.jpg",
  },
];

export function AvatarsShowcase() {
  const [activeAvatar, setActiveAvatar] = useState(avatars[0]);

  return (
    <section id="demo" className="py-16 lg:py-24 px-6 max-w-7xl mx-auto w-full">
      <div className="mb-12 text-center">
        <h2 className="text-4xl md:text-5xl font-outfit font-bold text-white mb-6 tracking-tight">
          Talk to your bot face-to-face.
        </h2>
        <p className="text-xl text-zinc-400 max-w-2xl mx-auto">
          Stop typing. Start talking with our hyper-realistic video avatars powered by Trugen. Give any OpenClaw configuration a custom identity.
        </p>
      </div>

      <div className="flex flex-col lg:flex-row gap-8 items-start">
        {/* Sidebar Selector */}
        <div className="w-full lg:w-1/3 flex flex-col gap-3">
          {avatars.map((avatar) => (
            <button
              key={avatar.id}
              onClick={() => setActiveAvatar(avatar)}
              className={`p-4 rounded-xl border text-left transition-all flex items-center justify-between ${
                activeAvatar.id === avatar.id 
                  ? "bg-white/10 border-[#00E3AA] shadow-[0_0_20px_rgba(0,227,170,0.15)]" 
                  : "bg-white/5 border-white/10 hover:bg-white/10"
              }`}
            >
              <div>
                <p className="text-white font-semibold text-lg">{avatar.name}</p>
                <p className="text-zinc-400 text-sm">{avatar.role}</p>
              </div>
              {activeAvatar.id === avatar.id && (
                <div className="w-8 h-8 rounded-full bg-[#00E3AA]/20 flex items-center justify-center">
                  <div className="w-3 h-3 rounded-full bg-[#00E3AA] animate-pulse" />
                </div>
              )}
            </button>
          ))}
        </div>

        {/* Video Preview */}
        <div className="w-full lg:w-2/3">
          <motion.div 
            key={activeAvatar.id}
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ duration: 0.4 }}
            className="w-full rounded-2xl overflow-hidden bg-[#0A0A0A] border border-white/10 relative shadow-2xl"
            style={{ aspectRatio: "16/9" }}
          >
            <Image
              src={activeAvatar.image}
              alt={activeAvatar.name}
              fill
              className="object-cover"
            />
            {/* Overlay UI */}
            <div className="absolute inset-0 bg-gradient-to-t from-black/80 via-transparent to-black/20" />
            
            <div className="absolute inset-x-0 bottom-0 p-6 flex items-end justify-between">
              <div className="flex items-center gap-4">
                <button className="w-14 h-14 rounded-full bg-[#00E3AA] hover:bg-[#00c996] transition-colors flex items-center justify-center text-black shadow-lg">
                  <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                    <path d="M23 7l-7 5 7 5V7z" /><rect x="1" y="5" width="15" height="14" rx="2" ry="2" />
                  </svg>
                </button>
                <div>
                  <p className="text-white font-semibold flex items-center gap-2">
                    {activeAvatar.name}
                  </p>
                  <p className="text-[#00E3AA] text-sm flex items-center gap-1.5">
                    <span className="w-2 h-2 rounded-full bg-[#00E3AA] animate-pulse" />
                    OpenClaw Agent Connected
                  </p>
                </div>
              </div>

              {/* Mock Chat bubbles */}
              <div className="hidden sm:flex flex-col gap-3 max-w-[300px] text-sm">
                <div className="bg-black/60 backdrop-blur-md text-white p-3 rounded-2xl rounded-tr-sm border border-white/10 self-end">
                  What's the status of the deploy pipeline?
                </div>
                <div className="bg-white/10 backdrop-blur-md text-white p-3 rounded-2xl rounded-tl-sm border border-white/10 self-start">
                  The production deploy is running. Health check passed on 2 of 3 pods.
                </div>
              </div>
            </div>
          </motion.div>
        </div>
      </div>
    </section>
  );
}
