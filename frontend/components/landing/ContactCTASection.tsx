"use client";

import { useState } from "react";
import { motion } from "framer-motion";

const INPUT_STYLE: React.CSSProperties = {
  background: "#1a1a1d",
  border: "1px solid rgba(255,255,255,0.08)",
  borderRadius: 12,
  color: "white",
  fontSize: 14,
  padding: "12px 16px",
  width: "100%",
  outline: "none",
};

export function ContactCTASection() {
  const [form, setForm] = useState({ firstName: "", lastName: "", email: "", description: "" });

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    setForm((prev) => ({ ...prev, [e.target.name]: e.target.value }));
  };

  return (
    <section id="contact" className="py-14 md:py-24 px-6 bg-canvas relative overflow-hidden">
      {/* Scattered dots */}
      {[
        { top: "8%",  left: "2%"  }, { top: "25%", left: "5%"  },
        { top: "55%", left: "3%"  }, { top: "78%", left: "7%"  },
        { top: "12%", left: "93%" }, { top: "40%", left: "96%" },
        { top: "68%", left: "94%" }, { top: "88%", left: "92%" },
        { top: "30%", left: "50%" }, { top: "60%", left: "55%" },
        { top: "45%", left: "20%" }, { top: "72%", left: "40%" },
      ].map((d, i) => (
        <div key={i} className="absolute w-1 h-1 rounded-full bg-white/10 pointer-events-none"
          style={{ top: d.top, left: d.left }} />
      ))}

      <div className="relative z-10 max-w-6xl mx-auto grid grid-cols-1 md:grid-cols-2 gap-14 items-center">
        {/* LEFT — copy */}
        <motion.div
          initial={{ opacity: 0, x: -24 }}
          whileInView={{ opacity: 1, x: 0 }}
          viewport={{ once: true }}
        >
          <div className="inline-flex items-center px-4 py-1.5 rounded-md bg-brand-subtle border border-brand-muted/20 mb-7">
            <span className="text-brand-muted text-sm font-mono font-semibold tracking-wide">
              Question &amp; Answers
            </span>
          </div>

          <h2 className="text-3xl md:text-[56px] font-bold leading-tight mb-6 capitalize">
            <span className="text-white">We are here</span>
            <br />
            <span className="text-danger">to help</span>
          </h2>

          <p className="text-body text-base md:text-lg leading-7 max-w-sm mb-10 md:mb-14">
            Have questions, ideas, or need help getting started? We&apos;re here
            for you. Reach out to our team and we&apos;ll get back to you as soon
            as possible.
          </p>

          <div>
            <p className="text-zinc-500 text-sm mb-1">Email us at</p>
            <p className="text-white font-semibold text-base">support@clawdface.ai</p>
          </div>
        </motion.div>

        {/* RIGHT — form */}
        <motion.div
          initial={{ opacity: 0, x: 24 }}
          whileInView={{ opacity: 1, x: 0 }}
          viewport={{ once: true }}
          className="rounded-3xl p-5 md:p-8"
          style={{ background: "#111113", border: "1px solid rgba(255,255,255,0.07)" }}
        >
          <div className="grid grid-cols-2 gap-4 mb-4">
            <div>
              <label className="text-white text-sm font-medium block mb-2">First Name</label>
              <input
                name="firstName"
                value={form.firstName}
                onChange={handleChange}
                placeholder="i.e john"
                style={INPUT_STYLE}
                className="placeholder-zinc-600 focus:border-brand/40 transition-colors"
              />
            </div>
            <div>
              <label className="text-white text-sm font-medium block mb-2">Last Name</label>
              <input
                name="lastName"
                value={form.lastName}
                onChange={handleChange}
                placeholder="i.e doe"
                style={INPUT_STYLE}
                className="placeholder-zinc-600 focus:border-brand/40 transition-colors"
              />
            </div>
          </div>

          <div className="mb-4">
            <label className="text-white text-sm font-medium block mb-2">Email Address</label>
            <input
              name="email"
              type="email"
              value={form.email}
              onChange={handleChange}
              placeholder="i.e johndoe@gmail.com"
              style={INPUT_STYLE}
              className="placeholder-zinc-600 focus:border-brand/40 transition-colors"
            />
          </div>

          <div className="mb-6">
            <label className="text-white text-sm font-medium block mb-2">Description</label>
            <textarea
              name="description"
              value={form.description}
              onChange={handleChange}
              placeholder="type your query here..."
              rows={7}
              style={{ ...INPUT_STYLE, resize: "none" }}
              className="placeholder-zinc-600 focus:border-brand/40 transition-colors"
            />
          </div>

          <button
            className="w-full py-4 rounded-2xl bg-brand text-black font-black text-sm tracking-widest uppercase hover:bg-brand-hover transition-colors"
          >
            Send Query
          </button>
        </motion.div>
      </div>
    </section>
  );
}
