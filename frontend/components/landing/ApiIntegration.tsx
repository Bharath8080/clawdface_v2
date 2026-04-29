import { motion } from "framer-motion";

export function ApiIntegration() {
  return (
    <section id="architecture" className="py-16 lg:py-24 px-6 max-w-7xl mx-auto w-full">
      <div className="flex flex-col lg:flex-row items-center gap-16">
        <div className="w-full lg:w-1/2">
          <h2 className="text-4xl md:text-5xl font-outfit font-bold text-white mb-6">
            The Stateless Bridge for Human-AI Interaction.
          </h2>
          <p className="text-xl text-zinc-400 mb-8 leading-relaxed">
            ClawdFace acts as a high-performance, stateless bridge. We handle the heavy lifting of WebRTC signaling, STT/TTS synchronization, and real-time avatar rendering, so you can focus on your core LLM logic.
          </p>
          
          <ul className="space-y-4 mb-8">
            {[
              "Drop-in Next.js components for voice/video UI",
              "100% Stateless 'Mega-Token' identification",
              "Direct connection to any OpenClaw-compatible endpoint",
              "Sub-300ms latency for real-time turn detection"
            ].map((item, i) => (
              <li key={i} className="flex items-center gap-3 text-zinc-300 font-medium">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#00E3AA" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
                  <polyline points="22 4 12 14.01 9 11.01" />
                </svg>
                {item}
              </li>
            ))}
          </ul>
          
          <button className="text-brand font-semibold hover:text-white transition-colors flex items-center gap-2">
            Read the Architecture Documentation
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <line x1="5" y1="12" x2="19" y2="12" />
              <polyline points="12 5 19 12 12 19" />
            </svg>
          </button>
        </div>

        <div className="w-full lg:w-1/2">
          <div className="bg-canvas rounded-xl border border-white/10 shadow-2xl overflow-hidden">
            <div className="flex items-center px-4 py-3 border-b border-white/5 bg-surface-secondary">
              <div className="flex gap-2">
                <div className="w-3 h-3 rounded-full bg-red-500/20 border border-red-500/50" />
                <div className="w-3 h-3 rounded-full bg-yellow-500/20 border border-yellow-500/50" />
                <div className="w-3 h-3 rounded-full bg-green-500/20 border border-green-500/50" />
              </div>
              <p className="ml-4 text-xs font-mono text-zinc-500">payload.json</p>
            </div>
            <div className="p-6 overflow-x-auto">
              <pre className="text-sm font-mono text-brand">
{`{
  "openclawUrl": "https://api.yourdomain.com",
  "gatewayToken": "YOUR_OPENCLAW_TOKEN",
  "avatarId": "priya_base_v1",
  "sessionKey": "session-12345",
  "instructions": "You are a professional AI assistant..."
}`}
              </pre>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
