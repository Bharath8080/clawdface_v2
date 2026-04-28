"use client";

import { motion } from "framer-motion";

const plans = [
  {
    name: "Free",
    price: "Free",
    description: "Perfect for testing and small projects",
    features: [
      { text: "10 minutes total usage", included: true },
      { text: "1 user concurrency", included: true },
      { text: "Intended for trial usage", included: true },
      { text: "Standard TTS/STT", included: true },
      { text: "Strict monthly limits", included: false },
    ],
    cta: "Get Started",
    popular: false,
  },
  {
    name: "Paid",
    price: "$9",
    subtext: "~ ₹900 / month",
    description: "For creators and growing businesses",
    features: [
      { text: "30 minutes max call duration", included: true },
      { text: "3-4 user concurrency", included: true },
      { text: "Priority STT/TTS (ElevenLabs/Deepgram)", included: true },
      { text: "Full avatar customization", included: true },
      { text: "Advanced turn detection", included: true },
    ],
    cta: "Upgrade to Pro",
    popular: true,
  },
  {
    name: "Enterprise",
    price: "Custom",
    description: "Fully configurable for high-scale clients",
    features: [
      { text: "Unlimited concurrency", included: true },
      { text: "Unlimited custom actions", included: true },
      { text: "Dedicated infrastructure", included: true },
      { text: "Custom white-label options", included: true },
      { text: "Fully configurable per client", included: true },
    ],
    cta: "Contact Sales",
    popular: false,
  },
];

export function PricingSection() {
  return (
    <section id="pricing" className="py-24 lg:py-32 px-6 bg-canvas relative overflow-hidden">
      {/* Background Glow */}
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[800px] h-[800px] bg-brand/5 rounded-full blur-[120px] pointer-events-none" />

      <div className="max-w-7xl mx-auto relative z-10">
        <div className="text-center mb-16">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
          >
            <h2 className="text-4xl md:text-6xl font-outfit font-bold text-white mb-6">
              Choose Your Plan
            </h2>
          </motion.div>
          <motion.p 
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.1 }}
            className="text-xl text-zinc-300 max-w-2xl mx-auto"
          >
            Flexible plans designed to scale with your AI interactions. <br className="hidden md:block" />
            No hidden fees, no complicated setup.
          </motion.p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
          {plans.map((plan, idx) => (
            <motion.div
              key={idx}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              transition={{ delay: idx * 0.1 }}
              className={`relative p-8 rounded-3xl border transition-all duration-300 ${
                plan.popular 
                  ? "bg-white/[0.05] border-brand/30 shadow-[0_0_40px_-15px_rgba(0,227,170,0.2)]" 
                  : "bg-white/[0.02] border-white/10"
              }`}
            >
              {plan.popular && (
                <div className="absolute -top-4 left-1/2 -translate-x-1/2 bg-brand text-black px-4 py-1 rounded-full text-xs font-bold uppercase tracking-wider">
                  Most Popular
                </div>
              )}

              <div className="mb-8">
                <h3 className="text-2xl font-bold text-white mb-2">{plan.name}</h3>
                <div className="flex items-baseline gap-1 mb-2">
                  <span className="text-5xl font-bold text-white">{plan.price}</span>
                  {plan.price !== "Custom" && plan.price !== "Free" && <span className="text-zinc-500">/month</span>}
                </div>
                {plan.subtext && (
                  <p className="text-sm text-brand font-medium mb-4">{plan.subtext}</p>
                )}
                <p className="text-zinc-300 text-sm">{plan.description}</p>
              </div>

              <div className="space-y-4 mb-10">
                {plan.features.map((feature, fIdx) => (
                  <div key={fIdx} className="flex items-start gap-3">
                    <div className={`mt-1 flex-shrink-0 rounded-full p-0.5 ${feature.included ? "text-brand" : "text-zinc-600"}`}>
                      {feature.included ? (
                        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round">
                          <polyline points="20 6 9 17 4 12" />
                        </svg>
                      ) : (
                        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round">
                          <line x1="18" y1="6" x2="6" y2="18" /><line x1="6" y1="6" x2="18" y2="18" />
                        </svg>
                      )}
                    </div>
                    <span className={`text-sm ${feature.included ? "text-zinc-200" : "text-zinc-500 line-through"}`}>
                      {feature.text}
                    </span>
                  </div>
                ))}
              </div>

              <button 
                className={`w-full py-4 rounded-xl font-bold transition-all ${
                  plan.popular
                    ? "bg-brand text-black hover:scale-[1.02] active:scale-95 shadow-lg shadow-brand/20"
                    : "bg-white/10 text-white hover:bg-white/20"
                }`}
              >
                {plan.cta}
              </button>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}
