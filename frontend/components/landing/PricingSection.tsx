"use client";

import Image from "next/image";
import { motion } from "framer-motion";
import { useRouter } from "next/navigation";

// ── Feature row ────────────────────────────────────────────────────────────
function Feature({ text }: { text: string }) {
  return (
    <div className="flex items-center gap-3">
      <Image src="/icons/healthicons_yes.svg" alt="check" width={20} height={20} className="shrink-0 object-contain" />
      <span className="text-zinc-300 text-sm">{text}</span>
    </div>
  );
}

const PLANS = [
  {
    name: "Free",
    icon: "/icons/Frame 18.svg",
    description: "Perfect for testing and small projects.",
    price: "Free",
    priceUnit: null,
    features: [
      "10 Free Minutes Monthly",
      "Full transcripts",
      "Community support",
      "5 min Session Limit",
      "1 Concurrent Session",
      "Support for 30+ Languages",
    ],
    cta: "Get Started",
    popular: false,
  },
  {
    name: "Pro",
    icon: "/icons/Frame 16.svg",
    description: "For individuals and small teams.",
    price: "$9",
    priceUnit: "/ per month",
    features: [
      "60 Free Minutes Monthly",
      "$0.15 per minute",
      "$0.12/min for additional usage",
      "4 Concurrent Sessions",
      "Limited Custom Avatars",
      "30 min Session Limit",
    ],
    cta: "Upgrade to Pro",
    popular: true,
  },
  {
    name: "Enterprise",
    icon: "/icons/Frame 18.svg",
    description: "For businesses that need more.",
    price: "Custom",
    priceUnit: null,
    features: [
      "Unlimited Concurrent Sessions",
      "Scaling Discounts",
      "Dev team integration support",
      "100% white-labeled experience",
      "Top tier customer support",
      "Enterprise grade security & compliance",
      "Guaranteed SLAs",
      "Dedicated isolated deployments",
    ],
    cta: "Contact Sales",
    popular: false,
  },
];

export function PricingSection() {
  const router = useRouter();

  const handleCta = (cta: string) => {
    if (cta === "Get Started" || cta === "Upgrade to Pro") {
      router.push("/sign-up");
    } else if (cta === "Contact Sales") {
      document.getElementById("contact")?.scrollIntoView({ behavior: "smooth" });
    }
  };

  return (
    <section id="pricing" className="py-14 md:py-24 px-6 bg-canvas relative overflow-hidden scroll-mt-24">
      {/* Ambient glow */}
      <div className="pointer-events-none absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[700px] h-[700px] bg-brand/5 rounded-full blur-[140px]" />

      {/* Scattered dots */}
      {[
        { top: "6%",  left: "2%"  }, { top: "20%", left: "5%"  },
        { top: "60%", left: "3%"  }, { top: "82%", left: "7%"  },
        { top: "8%",  left: "94%" }, { top: "35%", left: "96%" },
        { top: "70%", left: "95%" }, { top: "88%", left: "93%" },
      ].map((d, i) => (
        <div key={i} className="absolute w-1 h-1 rounded-full bg-white/10 pointer-events-none"
          style={{ top: d.top, left: d.left }} />
      ))}

      <div className="relative z-10 max-w-5xl mx-auto">
        {/* Header */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          className="text-center mb-16"
        >
          <div className="inline-flex items-center px-4 py-1.5 rounded-md bg-brand-subtle border border-brand-muted/20 mb-6">
            <span className="text-brand-muted text-sm font-mono font-semibold tracking-wide">
              Pricing &amp; Plans
            </span>
          </div>
          <h2 className="text-3xl md:text-[56px] font-bold text-white mb-5 capitalize">
            Choose Your Plan
          </h2>
          <p className="text-body text-base md:text-lg max-w-lg mx-auto leading-7">
            Flexible plans designed to scale with your AI interactions.
            <br />No hidden fees, no complicated setup.
          </p>
        </motion.div>

        {/* Cards */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
          {PLANS.map((plan, idx) => (
            <motion.div
              key={idx}
              initial={{ opacity: 0, y: 24 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ delay: idx * 0.1 }}
              className="flex flex-col rounded-3xl p-7 group transition-all duration-300 cursor-default"
              whileHover={{ y: -6 }}
              style={{
                background: plan.popular ? "#0b1a14" : "#111113",
                border: plan.popular
                  ? "1px solid rgba(0,227,170,0.2)"
                  : "1px solid rgba(255,255,255,0.07)",
                boxShadow: "none",
                transition: "box-shadow 0.3s ease, border-color 0.3s ease",
              }}
              onMouseEnter={e => {
                (e.currentTarget as HTMLDivElement).style.boxShadow = plan.popular
                  ? "0 0 40px rgba(0,227,170,0.12), 0 8px 32px rgba(0,0,0,0.4)"
                  : "0 0 24px rgba(255,255,255,0.04), 0 8px 32px rgba(0,0,0,0.4)";
                (e.currentTarget as HTMLDivElement).style.borderColor = plan.popular
                  ? "rgba(0,227,170,0.45)"
                  : "rgba(255,255,255,0.18)";
              }}
              onMouseLeave={e => {
                (e.currentTarget as HTMLDivElement).style.boxShadow = "none";
                (e.currentTarget as HTMLDivElement).style.borderColor = plan.popular
                  ? "rgba(0,227,170,0.2)"
                  : "rgba(255,255,255,0.07)";
              }}
            >
              <div className="mb-5">
                <Image src={plan.icon} alt="avatar" width={40} height={40} className="object-contain" />
              </div>

              {/* Name + description */}
              <h3 className="text-white text-xl font-bold mb-1 leading-snug">{plan.name}</h3>
              <p className="text-zinc-500 text-sm mb-6 leading-snug">{plan.description}</p>

              {/* Price */}
              <div className="flex items-baseline gap-2 mb-7">
                <span className="text-white font-black text-5xl leading-none">{plan.price}</span>
                {plan.priceUnit && (
                  <span className="text-zinc-500 text-sm">{plan.priceUnit}</span>
                )}
              </div>

              {/* Features */}
              <p className="text-white text-sm font-semibold mb-4">What you will get</p>
              <div className="space-y-3 flex-1">
                {plan.features.map((f, i) => (
                  <Feature key={i} text={f} />
                ))}
              </div>

              {/* CTA */}
              <button
                onClick={() => handleCta(plan.cta)}
                className={`mt-8 w-full py-3.5 rounded-2xl text-sm font-bold transition-all duration-200 active:scale-95 ${
                  plan.popular
                    ? "bg-brand text-black hover:bg-brand-hover hover:shadow-[0_0_20px_rgba(0,227,170,0.35)]"
                    : "text-white hover:bg-white/15 hover:shadow-[0_0_16px_rgba(255,255,255,0.06)]"
                }`}
                style={
                  !plan.popular
                    ? { background: "rgba(255,255,255,0.07)", border: "1px solid rgba(255,255,255,0.1)" }
                    : undefined
                }
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
