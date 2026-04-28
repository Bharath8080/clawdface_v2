"use client";

import { motion } from "framer-motion";
import Link from "next/link";

export function ContactCTASection() {
  return (
    <section id="contact" className="py-28 px-6 bg-surface-secondary relative overflow-hidden">
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[400px] bg-brand/5 rounded-full blur-[100px] pointer-events-none" />

      <div className="max-w-4xl mx-auto text-center relative z-10">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
        >
          <div className="inline-flex items-center gap-2 bg-brand/10 border border-brand/20 px-4 py-2 rounded-full mb-8">
            <span className="w-2 h-2 rounded-full bg-brand animate-pulse" />
            <span className="text-brand text-sm font-semibold">Support</span>
          </div>

          <h2 className="text-5xl sm:text-6xl md:text-7xl font-bold text-white leading-[1.05] mb-6">
            We are here <br />
            <span className="text-brand">to help.</span>
          </h2>

          <p className="text-xl text-zinc-400 mb-10 max-w-xl mx-auto leading-relaxed">
            Have questions about ClawdFace? Need help getting started? Our team is ready to help you
            give your bot a face.
          </p>

          <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
            <Link href="mailto:support@clawdface.ai">
              <button className="px-8 py-4 bg-brand text-black rounded-xl font-bold hover:bg-brand-hover transition-colors text-base">
                support@clawdface.ai
              </button>
            </Link>
            <Link href="/sign-up">
              <button className="px-8 py-4 bg-white/5 text-white border border-white/10 rounded-xl font-medium hover:bg-white/10 transition-colors text-base">
                Get Started Free
              </button>
            </Link>
          </div>
        </motion.div>
      </div>
    </section>
  );
}
