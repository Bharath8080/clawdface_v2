"use client";

import { motion } from "framer-motion";
import Image from "next/image";

export function HowItWorksSection() {
  // Sync timings
  const simpleQueryDur = 1.5;
  const complexQueryDur = 2;

  return (
    <section id="how-it-works" className="py-16 lg:py-24 relative overflow-hidden bg-[#050505] scroll-mt-24">
      <div className="max-w-[1400px] mx-auto px-6">
        <div className="text-left mb-20 max-w-2xl">
          <h2 className="text-4xl md:text-5xl font-bold text-white mb-4 tracking-tight">How it works</h2>
          <p className="text-xl text-zinc-400">
            Real-time execution across distributed AI nodes.
          </p>
        </div>

        <div className="relative flex flex-col items-center">
          <motion.div
            whileInView={{ 
              boxShadow: [
                "0 0 0px rgba(130,232,178,0)",
                "0 0 30px rgba(130,232,178,0.2)",
                "0 0 0px rgba(130,232,178,0)"
              ],
              borderColor: [
                "rgba(130,232,178,0.2)",
                "rgba(130,232,178,0.8)",
                "rgba(130,232,178,0.2)"
              ],
            }}
            transition={{ 
              duration: simpleQueryDur, 
              repeat: Infinity, 
              ease: "easeInOut",
              times: [0, 0.5, 1]
            }}
            viewport={{ once: true }}
            className="w-48 h-32 rounded-[24px] border-2 bg-zinc-900/60 flex flex-col items-center justify-center p-4 relative overflow-hidden z-20 border-zinc-800"
          >
            <div className="absolute inset-0 bg-[#82e8b2]/[0.02]" />
            <div className="w-12 h-12 relative mb-3">
              <Image src="/trugenai.png" alt="Trugen AI Logo" fill className="object-contain" />
            </div>
            <span className="font-bold text-white text-[15px] tracking-tight">AI Assistant</span>
          </motion.div>

          {/* VERTICAL CONNECTORS (UP/DOWN) */}
          <div className="h-12 lg:h-14 w-12 relative hidden md:block z-0">
            {/* Up Path (Request) */}
            <div className="absolute left-0 top-0 bottom-0 w-[1px] bg-zinc-800" />
            <div className="absolute left-0 top-0 bottom-0 w-[1px] overflow-hidden">
              <motion.div 
                className="absolute left-0 w-[2px] h-16 bg-gradient-to-t from-transparent via-[#82e8b2] to-transparent"
                animate={{ top: ["100%", "-100%"] }}
                transition={{ duration: simpleQueryDur, repeat: Infinity, ease: "linear" }}
              />
            </div>
            <div className="absolute left-0 top-0 text-zinc-600 -translate-x-[5.5px] -translate-y-1">
              <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor"><path d="M12 2L4 10h16L12 2z"/></svg>
            </div>

            {/* Down Path (Response) */}
            <div className="absolute right-0 top-0 bottom-0 w-[1px] bg-zinc-800" />
            <div className="absolute right-0 top-0 bottom-0 w-[1px] overflow-hidden">
              <motion.div 
                className="absolute right-0 w-[2px] h-16 bg-gradient-to-b from-transparent via-[#82e8b2] to-transparent"
                animate={{ top: ["-100%", "100%"] }}
                transition={{ duration: simpleQueryDur, repeat: Infinity, ease: "linear", delay: simpleQueryDur/2 }}
              />
            </div>
            <div className="absolute right-0 bottom-0 text-zinc-600 translate-x-[5.5px] translate-y-1 rotate-180">
              <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor"><path d="M12 2L4 10h16L12 2z"/></svg>
            </div>
          </div>

          {/* MIDDLE LEVEL: FLOWS */}
          <div className="flex flex-col md:flex-row items-center w-full justify-center">
            
            {/* USER MODULE */}
            <motion.div
              whileInView={{ 
                borderColor: [
                  "rgba(255,255,255,0.1)",
                  "rgba(130,232,178,0.7)",
                  "rgba(255,255,255,0.1)"
                ],
              }}
              transition={{ 
                duration: simpleQueryDur, 
                repeat: Infinity, 
                ease: "easeInOut",
                delay: simpleQueryDur * 0.75 
              }}
              viewport={{ once: true }}
              className="w-40 h-40 rounded-[28px] border bg-zinc-900/60 flex flex-col items-center justify-center p-4 relative z-20 border-zinc-800"
            >
              <div className="text-zinc-500 mb-4 scale-110">
                <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                  <rect x="2" y="3" width="20" height="14" rx="2" ry="2"></rect>
                  <line x1="2" y1="20" x2="22" y2="20"></line>
                </svg>
              </div>
              <span className="font-bold text-white text-[15px] tracking-tight">User</span>
            </motion.div>

            {/* CONNECTOR: USER <-> HUB */}
            <div className="hidden md:flex w-12 lg:w-16 flex-col justify-center gap-6 relative z-0">
              {/* Forward Path */}
              <div className="w-full h-[1px] bg-zinc-800 relative">
                <div className="absolute inset-0 overflow-hidden">
                  <motion.div 
                    className="absolute top-0 h-[1px] w-16 bg-gradient-to-r from-transparent via-[#82e8b2] to-transparent"
                    animate={{ left: ["-100%", "100%"] }}
                    transition={{ duration: simpleQueryDur, repeat: Infinity, ease: "linear" }}
                  />
                </div>
                <div className="absolute right-0 top-1/2 -translate-y-1/2 translate-x-1.5 rotate-90 text-zinc-600">
                   <svg width="10" height="10" viewBox="0 0 24 24" fill="currentColor"><path d="M12 2L4 10h16L12 2z"/></svg>
                </div>
              </div>
              {/* Return Path */}
              <div className="w-full h-[1px] bg-zinc-800 relative">
                <div className="absolute inset-0 overflow-hidden">
                  <motion.div 
                   className="absolute top-0 h-[1px] w-16 bg-gradient-to-l from-transparent via-[#82e8b2] to-transparent"
                   animate={{ right: ["-100%", "100%"] }}
                   transition={{ duration: simpleQueryDur, repeat: Infinity, ease: "linear", delay: simpleQueryDur/2 }}
                  />
                </div>
                <div className="absolute left-0 top-1/2 -translate-y-1/2 -translate-x-1.5 -rotate-90 text-zinc-600">
                   <svg width="10" height="10" viewBox="0 0 24 24" fill="currentColor"><path d="M12 2L4 10h16L12 2z"/></svg>
                </div>
              </div>
            </div>

            {/* HUB MODULE: CLAWDFACE */}
            <motion.div
              whileInView={{ 
                boxShadow: [
                  "0 0 20px rgba(130,232,178,0.1)",
                  "0 0 40px rgba(130,232,178,0.25)",
                  "0 0 20px rgba(130,232,178,0.1)"
                ],
                borderColor: [
                  "rgba(130,232,178,0.4)",
                  "rgba(130,232,178,0.9)",
                  "rgba(130,232,178,0.4)"
                ]
              }}
              transition={{ duration: simpleQueryDur/2, repeat: Infinity, ease: "easeInOut" }}
              viewport={{ once: true }}
              className="w-52 h-36 rounded-[28px] border-2 bg-zinc-900/90 flex flex-col items-center justify-center p-4 relative z-30 shadow-[0_0_40px_rgba(130,232,178,0.1)] border-zinc-700"
            >
              <div className="w-12 h-12 relative mb-3">
                <Image src="/openclaw.png" alt="ClawdFace Logo" fill className="object-contain" />
              </div>
              <span className="font-bold text-white text-[16px]">ClawdFace</span>
            </motion.div>

            {/* CONNECTOR: HUB <-> CLAWDBOT */}
            <div className="hidden md:flex w-14 lg:w-20 flex-col justify-center gap-6 relative z-0">
              {/* Complex Request Path */}
              <div className="w-full h-[0] border-t-[1.5px] border-dashed border-zinc-800 relative">
                <div className="absolute top-1/2 left-0 right-0 h-4 -translate-y-1/2 overflow-visible">
                  <motion.div 
                    className="absolute top-0 w-12 h-4"
                    animate={{ left: ["-20%", "100%"] }}
                    transition={{ duration: complexQueryDur, repeat: Infinity, ease: "linear" }}
                  >
                    <div className="absolute right-0 top-1/2 -translate-y-1/2 w-8 h-[2px] bg-gradient-to-r from-transparent to-[#a78bfa] blur-[0.5px]" />
                    <div className="absolute right-0 top-1/2 -translate-y-1/2 w-[6px] h-[6px] translate-x-1/2 bg-[#a78bfa] rounded-full shadow-[0_0_12px_3px_#a78bfa]" />
                  </motion.div>
                </div>
                <div className="absolute right-0 top-1/2 -translate-y-1/2 translate-x-1.5 rotate-90 text-zinc-700">
                   <svg width="10" height="10" viewBox="0 0 24 24" fill="currentColor"><path d="M12 2L4 10h16L12 2z"/></svg>
                </div>
              </div>
              
              {/* Outgoing Result Path from Clawdbot */}
              <div className="w-full h-[0] border-t-[1.5px] border-dashed border-zinc-800 relative">
                <div className="absolute top-1/2 left-0 right-0 h-4 -translate-y-1/2 overflow-visible">
                  <motion.div 
                    className="absolute top-0 w-12 h-4"
                    animate={{ right: ["-20%", "100%"] }}
                    transition={{ duration: complexQueryDur, repeat: Infinity, ease: "linear", delay: 1 }}
                  >
                    <div className="absolute left-0 top-1/2 -translate-y-1/2 w-8 h-[2px] bg-gradient-to-l from-transparent to-[#82e8b2] blur-[0.5px]" />
                    <div className="absolute left-0 top-1/2 -translate-y-1/2 w-[6px] h-[6px] -translate-x-1/2 bg-[#82e8b2] rounded-full shadow-[0_0_12px_3px_#82e8b2]" />
                  </motion.div>
                </div>
                <div className="absolute left-0 top-1/2 -translate-y-1/2 -translate-x-1.5 -rotate-90 text-zinc-700">
                   <svg width="10" height="10" viewBox="0 0 24 24" fill="currentColor"><path d="M12 2L4 10h16L12 2z"/></svg>
                </div>
              </div>
            </div>

            {/* CLAWDBOT MODULE */}
            <motion.div
              whileInView={{ 
                boxShadow: [
                  "0 0 0px rgba(167,139,250,0)",
                  "0 0 30px rgba(167,139,250,0.3)",
                  "0 0 0px rgba(167,139,250,0)"
                ],
                borderColor: [
                  "rgba(167,139,250,0.2)",
                  "rgba(167,139,250,0.8)",
                  "rgba(167,139,250,0.2)"
                ],
              }}
              transition={{ 
                duration: complexQueryDur, 
                repeat: Infinity, 
                ease: "easeInOut",
                delay: complexQueryDur * 0.25 
              }}
              viewport={{ once: true }}
              className="w-40 h-40 rounded-[28px] border-2 bg-zinc-900/60 flex flex-col items-center justify-center p-4 relative z-20 border-zinc-800"
            >
              <div className="w-14 h-14 relative mb-3">
                <Image src="/openclaw.png" alt="Clawdbot Logo" fill className="object-contain" />
              </div>
              <span className="font-bold text-white text-[15px] tracking-tight">Clawdbot</span>
            </motion.div>

          </div>
        </div>
      </div>
    </section>
  );
}
