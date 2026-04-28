"use client";

import Image from "next/image";
import { motion } from "framer-motion";

const CARDS = [
  {
    title: "Acts As Your Digital Agent",
    description:
      "Your AI becomes a real presence — a face people can connect with, not just a chat bubble they type into.",
    image: "https://assets.trugen.ai/images/avatarImages/priya-wide.jpg",
    name: "Priya",
  },
  {
    title: "Human Feels Personal",
    description:
      "Face-to-face interaction builds trust instantly. Users feel seen, heard, and understood — every conversation.",
    image: "https://assets.trugen.ai/images/avatarImages/chole-wide.jpeg",
    name: "Chloe",
  },
  {
    title: "Makes Your Agent Feel Real",
    description:
      "Move beyond text and voice. A lifelike avatar with natural expressions transforms your bot into a true companion.",
    image: "https://assets.trugen.ai/images/avatarImages/aman-wide.jpg",
    name: "Aman",
  },
];

export function WhyFaceMattersSection() {
  return (
    <section className="py-24 px-6 bg-canvas">
      <div className="max-w-7xl mx-auto">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          className="text-center mb-16"
        >
          <div className="inline-flex items-center gap-2 bg-white/5 border border-white/10 px-4 py-2 rounded-full mb-6">
            <span className="text-zinc-400 text-sm font-semibold">Why It Matters</span>
          </div>
          <h2 className="text-4xl md:text-5xl font-bold text-white mb-4">
            Why Face Matters
          </h2>
          <p className="text-xl text-zinc-400 max-w-2xl mx-auto">
            Text alone can't build connection. A face changes everything.
          </p>
        </motion.div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          {CARDS.map((card, idx) => (
            <motion.div
              key={idx}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ delay: idx * 0.1 }}
              className="group rounded-2xl border border-white/10 bg-surface-card overflow-hidden hover:border-white/20 transition-all duration-300"
            >
              {/* Avatar image */}
              <div className="relative aspect-[4/3] overflow-hidden">
                <Image
                  src={card.image}
                  alt={card.name}
                  fill
                  className="object-cover object-top group-hover:scale-105 transition-transform duration-500"
                  sizes="(max-width: 768px) 100vw, 33vw"
                />
                <div className="absolute inset-0 bg-gradient-to-t from-surface-card via-transparent to-transparent" />
                {/* Live indicator */}
                <div className="absolute top-3 right-3 flex items-center gap-1.5 bg-black/50 backdrop-blur-sm px-2.5 py-1 rounded-full border border-white/10">
                  <span className="w-1.5 h-1.5 rounded-full bg-brand animate-pulse" />
                  <span className="text-[10px] text-white font-semibold uppercase tracking-wide">Live</span>
                </div>
              </div>
              {/* Text content */}
              <div className="p-6">
                <h3 className="text-xl font-bold text-white mb-3">{card.title}</h3>
                <p className="text-zinc-400 leading-relaxed text-sm">{card.description}</p>
              </div>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}
