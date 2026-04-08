"use client";

import { useState, useEffect } from "react";
import { motion, AnimatePresence } from "framer-motion";
import Link from "next/link";
import Image from "next/image";
import { useUser } from "@stackframe/stack";

const AVATARS = [
  { url: "https://assets.trugen.ai/images/avatarImages/priya-wide.jpg", name: "Priya" },
  { url: "https://assets.trugen.ai/images/avatarImages/chole-wide.jpeg", name: "Chloe" },
  { url: "https://assets.trugen.ai/images/avatarImages/aman-wide.jpg", name: "Aman" },
  { url: "https://assets.trugen.ai/images/avatarImages/matt.jpeg", name: "Matt" },
  { url: "https://assets.trugen.ai/images/avatarImages/sameer-wide.jpeg", name: "Sameer" }
];

export function HeroSection() {
  const user = useUser();
  const [currentIdx, setCurrentIdx] = useState(0);
  const ROTATION_TIME = 3500; // 3.5 seconds

  useEffect(() => {
    const timer = setInterval(() => {
      setCurrentIdx((prev) => (prev + 1) % AVATARS.length);
    }, ROTATION_TIME);
    return () => clearInterval(timer);
  }, []);

  return (
    <section className="relative pt-24 pb-16 px-6 sm:pt-32 sm:pb-20 lg:pb-24 overflow-hidden bg-[#0a0a0a]">
      <div className="max-w-[1400px] mx-auto flex flex-col lg:flex-row items-center gap-16 lg:gap-8">
        
        {/* Left Column: Copy & CTAs */}
        <div className="w-full lg:w-1/2 flex flex-col items-start z-10 text-left">
          
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5 }}
            className="flex items-center gap-3 text-sm font-mono mb-8"
          >
            <span className="bg-[#1a2d24] text-[#82e8b2] px-3 py-1 rounded-md font-semibold">
              OpenClaw
            </span>
            <span className="text-zinc-500">→</span>
            <span className="text-zinc-400">Interactive Avatars</span>
          </motion.div>

          <motion.h1
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.1 }}
            className="text-5xl sm:text-7xl lg:text-8xl font-bold tracking-tight text-white mb-6 leading-[1.1]"
          >
            Give Your <span className="text-[#ff5c5c]">Clawdbot</span><br/> a Real Face.
          </motion.h1>

          <motion.p
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.2 }}
            className="text-lg text-zinc-400 mb-10 max-w-xl leading-relaxed"
          >
            Your bot handles text. We handle the face and voice. Install the skill, verify your API key, and call your Clawdbot like a video call. It sees you, hears you speak, reads the transcript, and replies through a lifelike avatar. You change nothing in your OpenClaw logic.
          </motion.p>

          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.3 }}
            className="flex flex-col sm:flex-row items-center gap-4 w-full sm:w-auto"
          >
            <Link href={user ? "/dashboard" : "/handler/sign-up"} className="w-full sm:w-auto">
              <button className="w-full sm:w-auto px-8 py-4 bg-[#82e8b2] text-black rounded-xl font-bold hover:bg-[#6bd69e] transition-colors">
                {user ? "Go to Dashboard" : "Connect your Clawdbot (OpenClaw)"}
              </button>
            </Link>
            <Link href="#how-it-works" className="w-full sm:w-auto">
              <button className="w-full sm:w-auto px-6 py-4 bg-[#141414] text-white border border-white/5 rounded-xl font-medium hover:bg-[#1f1f1f] transition-colors flex items-center justify-center gap-3">
                <span className="text-left leading-tight">
                  Watch It Work
                </span>
                <svg width="24" height="24" viewBox="0 0 24 24" fill="white" xmlns="http://www.w3.org/2000/svg">
                  <path d="M8 5v14l11-7z" />
                </svg>
              </button>
            </Link>
          </motion.div>


        </div>

        {/* Right Column: Video UI Mockups */}
        <div className="w-full lg:w-1/2 flex justify-center relative">
          <motion.div
            initial={{ opacity: 0, scale: 0.95, x: 20 }}
            animate={{ opacity: 1, scale: 1, x: 0 }}
            transition={{ duration: 0.7, delay: 0.2 }}
            className="w-full max-w-lg aspect-[4/5] rounded-[2rem] border border-white/10 bg-[#111111] overflow-hidden shadow-2xl relative p-6 flex flex-col justify-between"
          >
            {/* Top Bar UI */}
            <div className="flex justify-between items-start z-20">
              <div>
                <p className="text-xs font-semibold text-zinc-500 uppercase tracking-widest mb-1">Real-time Interaction</p>
                <p className="text-[#82e8b2] font-semibold text-lg tracking-tight">
                  Interacting with {AVATARS[currentIdx].name}
                </p>
              </div>
              <div className="flex gap-2">
                <div className="w-2.5 h-2.5 rounded-full bg-red-500/50" />
                <div className="w-2.5 h-2.5 rounded-full bg-yellow-500/50" />
                <div className="w-2.5 h-2.5 rounded-full bg-green-500/50" />
              </div>
            </div>

            {/* Simulated Avatar Area with Swipe Animation */}
            <div className="absolute top-28 bottom-32 inset-x-6 rounded-2xl overflow-hidden shadow-inner border border-white/5 bg-[#050505] group/mockup">
              <div className="relative w-full h-full bg-[#050505]">
                <AnimatePresence initial={false}>
                  <motion.div
                    key={currentIdx}
                    initial={{ x: "100%", opacity: 0 }}
                    animate={{ x: 0, opacity: 1 }}
                    exit={{ x: "-100%", opacity: 0 }}
                    transition={{ 
                      type: "spring", 
                      stiffness: 100, 
                      damping: 20, 
                      mass: 1 
                    }}
                    className="absolute inset-0 flex items-center justify-center"
                  >
                    <Image
                      src={AVATARS[currentIdx].url}
                      alt={AVATARS[currentIdx].name}
                      fill
                      className="object-cover"
                      sizes="(max-width: 768px) 100vw, 500px"
                      priority
                    />
                  </motion.div>
                </AnimatePresence>
              </div>

              {/* Progress Bar & Dots Container */}
              <div className="absolute bottom-0 inset-x-0 z-20">
                {/* Dots */}
                <div className="flex justify-center gap-1.5 mb-2">
                  {AVATARS.map((_, i) => (
                    <div 
                      key={i} 
                      className={`h-1 rounded-full transition-all duration-300 ${i === currentIdx ? 'w-4 bg-[#82e8b2]' : 'w-1 bg-white/20'}`} 
                    />
                  ))}
                </div>
                {/* Animated Progress Line */}
                <div className="h-0.5 w-full bg-white/5 overflow-hidden">
                  <motion.div
                    key={`progress-${currentIdx}`}
                    initial={{ scaleX: 0 }}
                    animate={{ scaleX: 1 }}
                    transition={{ duration: ROTATION_TIME / 1000, ease: "linear" }}
                    className="h-full bg-[#82e8b2] origin-left"
                  />
                </div>
              </div>
              
              {/* Visualizer wave simulation (Stays on top) */}
              <div className="absolute bottom-10 inset-x-0 flex items-center justify-center gap-1 p-2 z-10">
                {[...Array(24)].map((_, i) => (
                  <div key={i} className="w-1.5 bg-[#82e8b2] rounded-full animate-pulse shadow-[0_0_15px_rgba(130,232,178,0.4)]" style={{ height: `${10 + ((i * 17) % 40)}px`, animationDelay: `${i * 0.1}s` }} />
                ))}
              </div>
            </div>

            {/* Bottom Call Controls */}
            <div className="flex justify-center items-center gap-8 z-20 pt-4 pb-2">
               <button className="flex flex-col items-center gap-2 group/btn">
                 <div className="w-12 h-12 rounded-full bg-white/5 border border-white/10 flex items-center justify-center group-hover/btn:bg-white/10 transition-colors">
                   <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M12 2a3 3 0 0 0-3 3v7a3 3 0 0 0 6 0V5a3 3 0 0 0-3-3z"/><path d="M19 10v2a7 7 0 0 1-14 0v-2"/><line x1="12" y1="19" x2="12" y2="22"/></svg>
                 </div>
                 <span className="text-[10px] text-zinc-500 font-medium tracking-wide">Mic</span>
               </button>
               
               <button className="flex flex-col items-center gap-2 group/btn">
                 <div className="w-16 h-16 rounded-full bg-red-500/90 flex items-center justify-center hover:bg-red-500 transition-colors shadow-lg shadow-red-500/20">
                   <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M10.68 13.31a16 16 0 0 0 3.41 2.6l1.27-1.27a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7 2 2 0 0 1 1.72 2v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6 19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L8.09 9.91"/></svg>
                 </div>
                 <span className="text-[10px] text-zinc-500 font-bold uppercase tracking-widest">End</span>
               </button>

               <button className="flex flex-col items-center gap-2 group/btn">
                 <div className="w-12 h-12 rounded-full bg-white/5 border border-white/10 flex items-center justify-center group-hover/btn:bg-white/10 transition-colors">
                   <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><polygon points="23 7 16 12 23 17 23 7"/><rect x="1" y="5" width="15" height="14" rx="2" ry="2"/></svg>
                 </div>
                 <span className="text-[10px] text-zinc-500 font-medium tracking-wide">Video</span>
               </button>
            </div>
            
            <div className="text-center absolute bottom-4 inset-x-0 text-sm text-zinc-600 font-mono">02:45</div>
          </motion.div>
        </div>
      </div>
    </section>
  );
}
