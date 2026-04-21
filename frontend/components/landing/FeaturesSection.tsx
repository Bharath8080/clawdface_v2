import { motion } from "framer-motion";

const features = [
  {
    title: "Realistic Human-to-AI Interactions",
    description: "Experience hyper-realistic, human-like conversations with AI avatars that respond with natural expressions and fluid movement.",
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#00E3AA" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
        <circle cx="9" cy="7" r="4" />
        <path d="M23 21v-2a4 4 0 0 0-3-3.87" />
        <path d="M16 3.13a4 4 0 0 1 0 7.75" />
      </svg>
    ),
    colSpan: "col-span-1 md:col-span-2",
  },
  {
    title: "Precision Voice & Speed",
    description: "Engineered with Deepgram for sub-300ms Speech-to-Text and ElevenLabs for industry-leading, high-fidelity neural voices.",
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#00E3AA" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z" />
        <path d="M19 10v2a7 7 0 0 1-14 0v-2" />
        <line x1="12" y1="19" x2="12" y2="23" />
        <line x1="8" y1="23" x2="16" y2="23" />
      </svg>
    ),
    colSpan: "col-span-1 md:col-span-1",
  },
  {
    title: "Beyond Text & Audio",
    description: "Move past traditional text chatbots and voice-only assistants; engage with your AI through a life-like, immersive video presence.",
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#00E3AA" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M23 7l-7 5 7 5V7z" /><rect x="1" y="5" width="15" height="14" rx="2" ry="2" />
      </svg>
    ),
    colSpan: "col-span-1 md:col-span-1",
  },
  {
    title: "Immersive Web Interaction",
    description: "A state-of-the-art web platform built specifically for seamless, face-to-face avatar interactions without any complex setup.",
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#00E3AA" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <rect x="2" y="3" width="20" height="14" rx="2" ry="2" />
        <line x1="2" y1="10" x2="22" y2="10" />
        <path d="M7 21h10" />
        <path d="M12 17v4" />
      </svg>
    ),
    colSpan: "col-span-1 md:col-span-2",
  },
];

export function FeaturesSection() {
  return (
    <section id="features" className="py-24 px-6 relative overflow-hidden bg-[#0a0a0a]">
      <div className="mb-16 text-center max-w-3xl mx-auto">
        <div className="inline-flex items-center gap-2 bg-white/5 border border-white/10 px-4 py-2 rounded-full mb-6">
          <span className="text-zinc-400 text-sm font-semibold">Capabilities</span>
        </div>
        <h2 className="text-4xl md:text-5xl font-outfit font-bold text-white mb-6">
          Same OpenClaw.{" "}
          <span className="text-[#00E3AA]">Entirely New Experience.</span>
        </h2>
        <p className="text-xl text-zinc-400">
          ClawdFace bridges your custom LLM providers directly to LiveKit-powered interactive avatars.
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6 max-w-5xl mx-auto">
        {features.map((feature, idx) => (
          <div 
            key={idx} 
            className="p-8 rounded-2xl bg-white/5 border border-white/10 hover:border-white/20 transition-all group overflow-hidden relative"
          >
            <div className="absolute inset-0 bg-gradient-to-br from-[#00E3AA]/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity" />
            <div className="relative z-10">
              <div className="w-12 h-12 bg-[#00E3AA]/10 rounded-xl flex items-center justify-center mb-6">
                {feature.icon}
              </div>
              <h3 className="text-2xl font-bold text-white mb-3 tracking-tight">
                {feature.title}
              </h3>
              <p className="text-zinc-400 leading-relaxed">
                {feature.description}
              </p>
            </div>
          </div>
        ))}
      </div>
    </section>
  );
}
