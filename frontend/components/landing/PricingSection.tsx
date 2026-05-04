"use client";

import Image from "next/image";
import { motion } from "framer-motion";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { getPublicPricingPlans, type PlanType } from "@/app/services/pricingPaymentService";

interface DisplayPlan {
  name: string;
  icon: string;
  description: string;
  price: string;
  priceUnit: string | null;
  features: string[];
  cta: string;
  popular: boolean;
}

// ── Static config per plan type ────────────────────────────────────────────
function getPlanKey(slug: string): "free" | "pro" | "enterprise" | null {
  if (slug.startsWith("free")) return "free";
  if (slug.startsWith("pro")) return "pro";
  if (slug.startsWith("ente_ente")) return "enterprise";
  return null;
}

const PLAN_CONFIG = {
  free:       { icon: "/icons/Frame 18.svg", cta: "Get Started",    popular: false },
  pro:        { icon: "/icons/Frame 16.svg", cta: "Upgrade to Pro", popular: true  },
  enterprise: { icon: "/icons/Frame 18.svg", cta: "Contact Sales",  popular: false },
};

const PLAN_ORDER = ["free", "pro", "enterprise"] as const;

function mapApiPlan(plan: PlanType): DisplayPlan | null {
  const key = getPlanKey(plan.slug);
  if (!key) return null;
  const config = PLAN_CONFIG[key];

  const price =
    plan.billingType === "free"
      ? "Free"
      : plan.price > 0
      ? `$${plan.price}`
      : "Custom";

  const priceUnit =
    plan.price > 0 ? `/ per ${plan.billingCycle}` : null;

  const features = plan.details
    .slice()
    .sort((a, b) => (a.order ?? 0) - (b.order ?? 0))
    .flatMap((d) => d.features);

  return {
    name: plan.displayName,
    icon: config.icon,
    description: plan.description,
    price,
    priceUnit,
    features,
    cta: config.cta,
    popular: config.popular,
  };
}

// ── Feature row ────────────────────────────────────────────────────────────
function Feature({ text }: { text: string }) {
  return (
    <div className="flex items-center gap-3">
      <Image src="/icons/healthicons_yes.svg" alt="check" width={20} height={20} className="shrink-0 object-contain" />
      <span className="text-zinc-300 text-sm">{text}</span>
    </div>
  );
}

// ── Skeleton card ──────────────────────────────────────────────────────────
function SkeletonCard() {
  return (
    <div className="flex flex-col rounded-3xl p-7 animate-pulse"
      style={{ background: "#111113", border: "1px solid rgba(255,255,255,0.07)" }}>
      <div className="w-10 h-10 bg-white/5 rounded-lg mb-5" />
      <div className="h-5 w-24 bg-white/5 rounded mb-2" />
      <div className="h-3 w-40 bg-white/5 rounded mb-6" />
      <div className="h-12 w-20 bg-white/5 rounded mb-7" />
      <div className="h-3 w-28 bg-white/5 rounded mb-4" />
      <div className="space-y-3 flex-1">
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="h-3 bg-white/5 rounded" style={{ width: `${60 + i * 8}%` }} />
        ))}
      </div>
      <div className="mt-8 h-12 bg-white/5 rounded-2xl" />
    </div>
  );
}

// ── Section ────────────────────────────────────────────────────────────────
export function PricingSection() {
  const router = useRouter();
  const [plans, setPlans] = useState<DisplayPlan[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getPublicPricingPlans().then(({ data }) => {
      if (!data) return;
      const mapped: Record<string, DisplayPlan> = {};
      data.forEach((p) => {
        const display = mapApiPlan(p);
        if (display) {
          const key = getPlanKey(p.slug)!;
          mapped[key] = display;
        }
      });
      setPlans(PLAN_ORDER.map((k) => mapped[k]).filter(Boolean) as DisplayPlan[]);
    }).finally(() => setLoading(false));
  }, []);

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
          {loading
            ? [0, 1, 2].map((i) => <SkeletonCard key={i} />)
            : plans.map((plan, idx) => (
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

                  <h3 className="text-white text-xl font-bold mb-1 leading-snug">{plan.name}</h3>
                  <p className="text-zinc-500 text-sm mb-6 leading-snug">{plan.description}</p>

                  <div className="flex items-baseline gap-2 mb-7">
                    <span className="text-white font-black text-5xl leading-none">{plan.price}</span>
                    {plan.priceUnit && (
                      <span className="text-zinc-500 text-sm">{plan.priceUnit}</span>
                    )}
                  </div>

                  <p className="text-white text-sm font-semibold mb-4">What you will get</p>
                  <div className="space-y-3 flex-1">
                    {plan.features.map((f, i) => (
                      <Feature key={i} text={f} />
                    ))}
                  </div>

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
