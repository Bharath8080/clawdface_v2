"use client";

import { motion } from "framer-motion";

const BADGES = [
  {
    title: "GDPR Compliant",
    description: "Full EU data protection compliance. User data is never stored beyond the session.",
    icon: (
      <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="#00E3AA" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
      </svg>
    ),
    accent: "border-[#00E3AA]/20 bg-[#00E3AA]/5",
    iconBg: "bg-[#00E3AA]/10",
  },
  {
    title: "ISO Compliant",
    description: "Meets ISO 27001 international standards for information security management.",
    icon: (
      <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="#60a5fa" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="12" cy="12" r="10" />
        <polyline points="12 6 12 12 16 14" />
      </svg>
    ),
    accent: "border-blue-500/20 bg-blue-500/5",
    iconBg: "bg-blue-500/10",
  },
  {
    title: "HIPAA Compliant",
    description: "Healthcare-ready infrastructure with end-to-end encryption for sensitive data.",
    icon: (
      <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="#c084fc" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z" />
      </svg>
    ),
    accent: "border-purple-500/20 bg-purple-500/5",
    iconBg: "bg-purple-500/10",
  },
  {
    title: "AICPA SOC 2 Compliant",
    description: "Audited for security, availability, and confidentiality of customer data.",
    icon: (
      <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="#fbbf24" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
        <path d="M7 11V7a5 5 0 0 1 10 0v4" />
      </svg>
    ),
    accent: "border-yellow-500/20 bg-yellow-500/5",
    iconBg: "bg-yellow-500/10",
  },
];

export function SecuritySection() {
  return (
    <section className="py-24 px-6 bg-[#050505]">
      <div className="max-w-7xl mx-auto">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          className="text-center mb-16"
        >
          <div className="inline-flex items-center gap-2 bg-[#00E3AA]/10 border border-[#00E3AA]/20 px-4 py-2 rounded-full mb-6">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#00E3AA" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
            </svg>
            <span className="text-[#00E3AA] text-sm font-semibold">Enterprise Ready</span>
          </div>
          <h2 className="text-4xl md:text-5xl font-bold text-white mb-4">
            Enterprise Grade Security
          </h2>
          <p className="text-xl text-zinc-400 max-w-2xl mx-auto">
            Built for organizations that can&apos;t compromise on data protection or compliance.
          </p>
        </motion.div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
          {BADGES.map((badge, idx) => (
            <motion.div
              key={idx}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ delay: idx * 0.1 }}
              className={`rounded-2xl border p-6 ${badge.accent} hover:scale-[1.02] transition-transform duration-300`}
            >
              <div className={`w-14 h-14 rounded-2xl ${badge.iconBg} flex items-center justify-center mb-5`}>
                {badge.icon}
              </div>
              <h3 className="text-white font-bold text-lg mb-2">{badge.title}</h3>
              <p className="text-zinc-400 text-sm leading-relaxed">{badge.description}</p>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}
