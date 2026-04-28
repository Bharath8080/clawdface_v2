"use client";

import { motion } from "framer-motion";
import React, { useContext, useState } from "react";
import { SubscriptionContext, SubscriptionContextProvider } from "@/app/context/subscriptionContext";
import { ILicenseInfo, PlanType, updateAutoReload } from "@/app/services/pricingPaymentService";
import usePayment from "@/app/hooks/usePayment";
import { ContactUsModal } from "@/components/ContactUsModal";

// ─── Icons ──────────────────────────────────────────────────────────────────
const CheckIcon = () => (
  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
    <polyline points="20 6 9 17 4 12" />
  </svg>
);

const ZapIcon = ({ size = 18 }: { size?: number }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2" />
  </svg>
);

const CrownIcon = () => (
  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <path d="m2 4 3 12h14l3-12-6 7-4-7-4 7-6-7zm3 16h14" />
  </svg>
);

const BuildingIcon = () => (
  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <rect width="16" height="20" x="4" y="2" rx="2" ry="2" />
    <path d="M9 22v-4h6v4" />
    <path d="M8 6h.01M16 6h.01M12 6h.01M12 10h.01M12 14h.01M16 10h.01M16 14h.01M8 10h.01M8 14h.01" />
  </svg>
);

const CreditStackIcon = () => (
  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <ellipse cx="12" cy="5" rx="9" ry="3" />
    <path d="M3 5v4c0 1.66 4.03 3 9 3s9-1.34 9-3V5" />
    <path d="M3 13v4c0 1.66 4.03 3 9 3s9-1.34 9-3v-4" />
  </svg>
);

const SpinnerIcon = () => (
  <svg className="animate-spin" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
    <path d="M21 12a9 9 0 1 1-6.219-8.56" />
  </svg>
);

const ArrowRightIcon = () => (
  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
    <path d="M5 12h14M12 5l7 7-7 7" />
  </svg>
);

const PencilIcon = () => (
  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.4" strokeLinecap="round" strokeLinejoin="round">
    <path d="M12 20h9" />
    <path d="M16.5 3.5a2.12 2.12 0 0 1 3 3L7 19l-4 1 1-4Z" />
  </svg>
);

// ─── Helpers ─────────────────────────────────────────────────────────────────
function formatCredits(n: number): string {
  if (n >= 1000 && n % 1000 === 0) return `${n / 1000}K`;
  if (n >= 1000) return n.toLocaleString();
  return String(n);
}

function formatBillingDate(date?: string | null): string {
  if (!date) return "Not available";
  const parsed = new Date(date);
  if (Number.isNaN(parsed.getTime())) return "Not available";
  return parsed.toLocaleString(undefined, {
    day: "numeric",
    month: "short",
    year: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });
}

function getPlanMeta(plan: PlanType): {
  accent: string;
  icon: React.ReactNode;
  popular: boolean;
  isEnterprise: boolean;
  isFree: boolean;
} {
  const slug = plan.slug ?? "";
  if (slug.includes("free") || plan.billingType === "free") {
    return { accent: "#9ca3af", icon: <ZapIcon />, popular: false, isEnterprise: false, isFree: true };
  }
  if (slug.includes("ente_ente") || plan.name === "Enterprise") {
    return { accent: "#a78bfa", icon: <BuildingIcon />, popular: false, isEnterprise: true, isFree: false };
  }
  // Pro (or any paid subscription)
  return { accent: "#00E3AA", icon: <CrownIcon />, popular: true, isEnterprise: false, isFree: false };
}

function getAllFeatures(plan: PlanType): string[] {
  if (!Array.isArray(plan.details)) return [];
  return plan.details.flatMap((d) => d.features ?? []);
}

// ─── Subscription Plan Card ──────────────────────────────────────────────────
function PlanCard({
  plan,
  isCurrent,
  onUpgrade,
  onContactSales,
  isLoading,
  index,
}: {
  plan: PlanType;
  isCurrent: boolean;
  onUpgrade: (id: string) => void;
  onContactSales: () => void;
  isLoading: boolean;
  index: number;
}) {
  const { accent, icon, popular, isEnterprise, isFree } = getPlanMeta(plan);
  const features = getAllFeatures(plan);

  const handleClick = () => {
    if (isCurrent || isLoading) return;
    if (isEnterprise) { onContactSales(); return; }
    onUpgrade(plan.slug);
  };

  const ctaLabel = () => {
    if (isCurrent) return "Current Plan";
    if (isEnterprise) return "Contact Sales";
    if (isFree) return "Downgrade to Free";
    return `Upgrade to ${plan.displayName}`;
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 24 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: index * 0.08, duration: 0.4, ease: "easeOut" }}
      className={`relative flex flex-col p-7 rounded-2xl border transition-all duration-300
        ${popular
          ? "bg-white/[0.05] border-[#00E3AA]/30 shadow-[0_0_48px_-12px_rgba(0,227,170,0.2)]"
          : "bg-white/[0.02] border-white/8 hover:border-white/15 hover:bg-white/[0.03]"
        }`}
    >
      {/* Popular badge */}
      {popular && !isCurrent && (
        <div className="absolute -top-3.5 left-1/2 -translate-x-1/2 bg-[#00E3AA] text-black px-4 py-1 rounded-full text-[11px] font-bold uppercase tracking-wider whitespace-nowrap">
          Most Popular
        </div>
      )}

      {/* Current plan badge */}
      {isCurrent && (
        <div className="absolute -top-3.5 left-1/2 -translate-x-1/2 bg-white/10 text-white/60 border border-white/15 px-4 py-1 rounded-full text-[11px] font-semibold uppercase tracking-wider whitespace-nowrap">
          Current Plan
        </div>
      )}

      {/* Header */}
      <div className="mb-5">
        <div className="w-10 h-10 rounded-xl flex items-center justify-center mb-4"
          style={{ backgroundColor: `${accent}15`, color: accent }}>
          {icon}
        </div>
        <h3 className="text-xl font-bold text-white mb-1">{plan.displayName}</h3>
        <div className="flex items-baseline gap-1.5 mb-2">
          {isEnterprise ? (
            <span className="text-3xl font-bold text-white">Custom</span>
          ) : (
            <>
              <span className="text-4xl font-bold text-white tracking-tight">
                {plan.price === 0 ? "Free" : `$${plan.price}`}
              </span>
              {plan.price > 0 && (
                <span className="text-[#6b7280] text-sm">/ month</span>
              )}
            </>
          )}
        </div>
        <p className="text-[#9ca3af] text-sm leading-relaxed">{plan.description}</p>
      </div>

      {/* Stats row */}
      {(plan.credits > 0 || plan.concurrentSessions > 0) && (
        <div className="grid grid-cols-3 divide-x divide-white/[0.06] mb-5 py-3 rounded-xl bg-white/[0.03] border border-white/5">
          <div className="flex flex-col items-center justify-center gap-0.5 px-2 text-center">
            <span className="text-white font-bold text-sm leading-none">
              {plan.credits > 0 ? formatCredits(plan.credits) : "—"}
            </span>
            <span className="text-[#6b7280] text-[10px] mt-1 leading-tight">credits / mo</span>
          </div>
          <div className="flex flex-col items-center justify-center gap-0.5 px-2 text-center">
            <span className="text-white font-bold text-sm leading-none">
              {plan.concurrentSessions > 0
                ? plan.concurrentSessions >= 60 ? "∞" : plan.concurrentSessions
                : "—"}
            </span>
            <span className="text-[#6b7280] text-[10px] mt-1 leading-tight">concurrent</span>
          </div>
          <div className="flex flex-col items-center justify-center gap-0.5 px-2 text-center">
            <span className="text-white font-bold text-sm leading-none">
              {plan.maxSessionDuration > 0
                ? plan.maxSessionDuration >= 120 ? "∞" : `${plan.maxSessionDuration}m`
                : "—"}
            </span>
            <span className="text-[#6b7280] text-[10px] mt-1 leading-tight">per session</span>
          </div>
        </div>
      )}

      {/* Divider */}
      <div className="border-t border-white/5 mb-4" />

      {/* Features */}
      {features.length > 0 && (
        <div className="flex-1 space-y-2.5 mb-6">
          {features.map((f, i) => (
            <div key={i} className="flex items-start gap-2.5">
              <span className="mt-0.5 shrink-0" style={{ color: accent }}>
                <CheckIcon />
              </span>
              <span className="text-sm text-[#d1d5db] leading-snug">{f}</span>
            </div>
          ))}
        </div>
      )}

      {/* CTA */}
      <button
        onClick={handleClick}
        disabled={isCurrent || isLoading}
        className={`mt-auto w-full py-3.5 rounded-xl font-semibold text-[15px] transition-all duration-200 flex items-center justify-center gap-2
          ${isCurrent
            ? "bg-white/5 text-white/30 cursor-not-allowed border border-white/8"
            : isEnterprise
              ? "bg-[#a78bfa]/15 text-[#a78bfa] border border-[#a78bfa]/25 hover:bg-[#a78bfa]/25 active:scale-[0.98]"
              : isFree
                ? "bg-white/8 text-white border border-white/10 hover:bg-white/12 active:scale-[0.98]"
                : "bg-[#00E3AA] text-black hover:bg-[#00E3AA]/90 active:scale-[0.98] shadow-lg shadow-[#00E3AA]/15"
          }`}
      >
        {isLoading && !isCurrent ? <SpinnerIcon /> : null}
        {ctaLabel()}
        {!isCurrent && isEnterprise && <ArrowRightIcon />}
      </button>
    </motion.div>
  );
}

// ─── Credit Top-up Card ───────────────────────────────────────────────────────
function CreditCard({
  plan,
  onBuy,
  isLoading,
  index,
}: {
  plan: PlanType;
  onBuy: (id: string) => void;
  isLoading: boolean;
  index: number;
}) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 16 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: index * 0.06, duration: 0.35, ease: "easeOut" }}
      className="flex flex-col items-center gap-4 p-5 rounded-2xl bg-white/[0.02] border border-white/8 hover:border-[#00E3AA]/20 hover:bg-white/[0.03] transition-all duration-300 group"
    >
      <div className="w-10 h-10 rounded-xl bg-[#00E3AA]/10 border border-[#00E3AA]/20 flex items-center justify-center text-[#00E3AA]">
        <CreditStackIcon />
      </div>
      <div className="text-center">
        <div className="text-2xl font-bold text-white tracking-tight">
          {formatCredits(plan.credits)}
        </div>
        <div className="text-xs text-[#6b7280] mt-0.5">credits</div>
      </div>
      <div className="text-[11px] text-[#4b5563] uppercase tracking-wider font-medium">
        One-time
      </div>
      <button
        onClick={() => onBuy(plan.slug)}
        disabled={isLoading}
        className="w-full py-2.5 rounded-xl text-sm font-semibold bg-[#00E3AA]/10 text-[#00E3AA] border border-[#00E3AA]/20 hover:bg-[#00E3AA]/20 active:scale-[0.97] transition-all duration-200 flex items-center justify-center gap-1.5 disabled:opacity-50"
      >
        {isLoading ? <SpinnerIcon /> : null}
        Add Credits
      </button>
    </motion.div>
  );
}

// ─── Skeleton loader ──────────────────────────────────────────────────────────
function PlanDetailsPanel({
  licenseInfo,
  currentPlanName,
  onAutoReloadChange,
  autoReloadLoading,
  autoReloadError,
}: {
  licenseInfo: ILicenseInfo;
  currentPlanName: string;
  onAutoReloadChange: (enabled: boolean) => void;
  autoReloadLoading: boolean;
  autoReloadError: string | null;
}) {
  const isAutoReloadOn = !!licenseInfo?.autoReload;
  const autoReloadSlug = licenseInfo?.autoReloadSlug || "20k_one_time";

  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: 0.08, duration: 0.4 }}
      className="mb-10 border border-white/15 bg-white/[0.035] px-6 py-5"
    >
      <div className="mb-4">
        <h2 className="text-[15px] font-bold text-white tracking-tight">Plan details</h2>
        <p className="mt-1 text-[12px] text-[#9ca3af]">
          Renews on {formatBillingDate(licenseInfo?.expiresAt)}
        </p>
      </div>

      <div className="grid grid-cols-1 divide-y divide-white/10 border-t border-white/10 pt-5 md:grid-cols-3 md:divide-x md:divide-y-0">
        <div className="pb-5 md:pb-0 md:pr-7">
          <div className="mb-3 flex items-center gap-2">
            <span className="text-[14px] font-semibold text-white">Current Plan</span>
            <span className="rounded-full bg-[#00E3AA] px-3 py-1 text-[11px] font-bold text-black">
              {currentPlanName}
            </span>
          </div>
          <p className="text-[12px] text-[#9ca3af]">Upgrade to get better rates</p>
        </div>

        <div className="py-5 md:px-7 md:py-0">
          <h3 className="mb-4 text-[14px] font-semibold text-white">Billing cycle usage</h3>
          <div className="space-y-3 text-[12px]">
            <div className="flex items-center justify-between gap-4">
              <span className="text-[#9ca3af]">Credits available</span>
              <span className="font-medium text-white/90">
                {formatCredits(licenseInfo?.balanceCredit ?? 0)}/{formatCredits(licenseInfo?.totalCredit ?? 0)}
              </span>
            </div>
            <div className="flex items-center justify-between gap-4">
              <span className="text-[#9ca3af]">Add-on credits</span>
              <span className="font-medium text-white/90">
                {formatCredits(licenseInfo?.purchasedCredit ?? 0)}
              </span>
            </div>
          </div>
        </div>

        <div className="pt-5 md:pl-7 md:pt-0">
          <div className="mb-3 flex items-start justify-between gap-4">
            <h3 className="text-[14px] font-semibold text-white">Auto-reload credits</h3>
            <button
              type="button"
              role="switch"
              aria-checked={isAutoReloadOn}
              disabled={autoReloadLoading}
              onClick={() => onAutoReloadChange(!isAutoReloadOn)}
              className={`relative h-6 w-11 shrink-0 rounded-full border transition-all duration-200 disabled:opacity-60 ${
                isAutoReloadOn
                  ? "border-[#00E3AA] bg-[#00E3AA]"
                  : "border-white/15 bg-white/10"
              }`}
            >
              <span
                className={`absolute left-0.5 top-0.5 h-5 w-5 rounded-full bg-white shadow-sm transition-transform duration-200 ${
                  isAutoReloadOn ? "translate-x-5" : "translate-x-0"
                }`}
              />
            </button>
          </div>
          <p className="text-[12px] leading-relaxed text-[#9ca3af]">
            {autoReloadSlug.replace(/_/g, " ")} will be reloaded automatically as add-on credits when the balance goes below 2000.
          </p>
          <button
            type="button"
            onClick={() => onAutoReloadChange(true)}
            disabled={autoReloadLoading}
            className="mt-3 inline-flex items-center gap-1.5 text-[12px] font-semibold text-[#00E3AA] hover:text-[#00ffd0] disabled:opacity-60"
          >
            {autoReloadLoading ? <SpinnerIcon /> : <PencilIcon />}
            Manage auto-reload
          </button>
          {autoReloadError && (
            <p className="mt-2 text-[12px] text-red-400">{autoReloadError}</p>
          )}
        </div>
      </div>
    </motion.div>
  );
}

function PlanSkeleton() {
  return (
    <div className="flex flex-col p-7 rounded-2xl bg-white/[0.02] border border-white/8 gap-4 animate-pulse">
      <div className="w-10 h-10 rounded-xl bg-white/5" />
      <div className="h-6 w-24 rounded-lg bg-white/5" />
      <div className="h-10 w-20 rounded-lg bg-white/5" />
      <div className="space-y-2 mt-2">
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="h-4 rounded-md bg-white/5" style={{ width: `${60 + i * 8}%` }} />
        ))}
      </div>
      <div className="mt-auto h-12 rounded-xl bg-white/5" />
    </div>
  );
}

// ─── Main inner component ─────────────────────────────────────────────────────
function SubscriptionViewInner() {
  const { allPricingPlans, licenseInfo, setLicenseInfo, isLoading } = useContext(SubscriptionContext);
  const { paymentHandler, loading: paymentLoading, error: paymentError } = usePayment();
  const [contactModalOpen, setContactModalOpen] = useState(false);
  const [autoReloadLoading, setAutoReloadLoading] = useState(false);
  const [autoReloadError, setAutoReloadError] = useState<string | null>(null);

  // Separate plan types from real API data
  const currentSlug = licenseInfo?.slug ?? "";
  const isPaidPlan = currentSlug && !currentSlug.includes("free");

  const subscriptionPlans = allPricingPlans
    .filter((p) => {
      if (p.billingCycle !== "month") return false;
      // Hide free plan when user is already on a paid plan
      if (isPaidPlan && (p.billingType === "free" || p.slug?.includes("free"))) return false;
      return true;
    })
    .sort((a, b) => {
      const order = (p: PlanType) => {
        if (p.billingType === "free" || p.slug?.includes("free")) return 0;
        if (p.slug?.includes("ente_ente") || p.name === "Enterprise") return 2;
        return 1;
      };
      return order(a) - order(b);
    });

  const creditPlans = allPricingPlans
    .filter((p) => p.billingType === "one-time")
    .sort((a, b) => a.credits - b.credits);

  const currentPlanName = subscriptionPlans.find((p) => p.slug === currentSlug)?.displayName
    ?? allPricingPlans.find((p) => p.slug === currentSlug)?.displayName
    ?? subscriptionPlans.find((p) => p.billingType === "free")?.displayName
    ?? "Free";

  const handleAutoReloadChange = async (enabled: boolean) => {
    const apiKey = typeof window !== "undefined"
      ? localStorage.getItem("defaultApiKey")
      : null;

    if (!apiKey) {
      setAutoReloadError("No API key found. Please refresh the page and try again.");
      return;
    }

    const autoReloadSlug = licenseInfo?.autoReloadSlug || "20k_one_time";

    setAutoReloadLoading(true);
    setAutoReloadError(null);
    const previous = licenseInfo;
    setLicenseInfo({ ...licenseInfo, autoReload: enabled, autoReloadSlug });

    const { error } = await updateAutoReload(apiKey, {
      autoReload: enabled,
      autoReloadSlug,
    });

    if (error) {
      setLicenseInfo(previous);
      setAutoReloadError(error);
    }

    setAutoReloadLoading(false);
  };

  return (
    <div className="h-full overflow-y-auto bg-[#050505] custom-scrollbar">
      {/* Background glows */}
      <div className="pointer-events-none fixed inset-0 overflow-hidden">
        <div className="absolute top-1/3 left-1/2 -translate-x-1/2 w-[600px] h-[600px] bg-[#00E3AA]/4 rounded-full blur-[120px]" />
        <div className="absolute top-2/3 right-1/4 w-[300px] h-[300px] bg-[#a78bfa]/3 rounded-full blur-[100px]" />
      </div>

      <div className="relative z-10 max-w-5xl mx-auto px-6 py-10 pb-16">

        {/* Header */}
        <motion.div initial={{ opacity: 0, y: 16 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.4 }} className="mb-8">
          <div className="flex items-center gap-2 mb-3">
            <span className="text-[11px] font-bold uppercase tracking-widest text-[#00E3AA] bg-[#00E3AA]/10 px-3 py-1 rounded-full border border-[#00E3AA]/20">
              Billing
            </span>
          </div>
          <h1 className="text-3xl font-bold text-white font-outfit mb-2 tracking-tight">
            {isLoading ? "Subscription" : isPaidPlan ? `You're on ${currentPlanName}` : "Choose Your Plan"}
          </h1>
          <p className="text-[#6b7280] text-[15px] max-w-2xl">
            {isLoading
              ? "Loading your subscription..."
              : isPaidPlan
                ? `Your ${currentPlanName} plan is active. Upgrade, add credits, or manage your subscription below.`
                : "Simple, transparent pricing that grows with you. Try free plan with no commitments."}
          </p>
        </motion.div>

        {/* Payment error */}
        {paymentError && (
          <motion.div
            initial={{ opacity: 0, y: -8 }}
            animate={{ opacity: 1, y: 0 }}
            className="mb-6 p-4 rounded-xl bg-red-500/10 border border-red-500/20 text-red-400 text-sm"
          >
            {paymentError}
          </motion.div>
        )}

        {/* ── Subscription plan cards ── */}
        {isPaidPlan && !isLoading && (
          <PlanDetailsPanel
            licenseInfo={licenseInfo}
            currentPlanName={currentPlanName}
            onAutoReloadChange={handleAutoReloadChange}
            autoReloadLoading={autoReloadLoading}
            autoReloadError={autoReloadError}
          />
        )}

        <motion.h2
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.1 }}
          className="text-[13px] text-center font-semibold text-[#6b7280] uppercase tracking-widest mb-5"
        >
          Subscription Plans
        </motion.h2>

        <div className={`grid gap-5 mb-12 grid-cols-1 ${
          subscriptionPlans.length === 1
            ? "max-w-sm mx-auto"
            : subscriptionPlans.length === 2
              ? "md:grid-cols-2 max-w-2xl mx-auto"
              : "md:grid-cols-3"
        }`}>
          {isLoading
            ? [0, 1, 2].map((i) => <PlanSkeleton key={i} />)
            : subscriptionPlans.map((plan, idx) => (
                <PlanCard
                  key={plan.id}
                  plan={plan}
                  index={idx}
                  isCurrent={plan.slug === currentSlug || (!currentSlug && plan.billingType === "free")}
                  onUpgrade={paymentHandler}
                  isLoading={paymentLoading}
                  onContactSales={() => setContactModalOpen(true)}
                />
              ))}
        </div>

        {/* ── Credit top-ups (Pro users only) ── */}
        {isPaidPlan && creditPlans.length > 0 && (
          <>
            <motion.div
              initial={{ opacity: 0, y: 12 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.2, duration: 0.4 }}
              className="mb-5"
            >
              <div className="flex items-center gap-3 mb-1">
                <h2 className="text-[13px] font-semibold text-[#6b7280] uppercase tracking-widest">
                  Add Credits
                </h2>
                <span className="text-[10px] font-bold text-[#00E3AA] bg-[#00E3AA]/10 border border-[#00E3AA]/20 px-2 py-0.5 rounded-full uppercase tracking-wider">
                  Pro
                </span>
              </div>
              <p className="text-[#4b5563] text-sm">
                Top up your account with additional credits — no subscription change required.
              </p>
            </motion.div>

            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-5 gap-4 mb-10">
              {creditPlans.map((plan, idx) => (
                <CreditCard
                  key={plan.id}
                  plan={plan}
                  index={idx}
                  onBuy={paymentHandler}
                  isLoading={paymentLoading}
                />
              ))}
            </div>
          </>
        )}

        {/* Manage subscription */}
        {licenseInfo?.stripeId && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.3 }}
            className="mb-8 text-center"
          >
            <button
              onClick={async () => {
                const apiKey = localStorage.getItem("defaultApiKey");
                if (!apiKey) return;
                const { getManageSubUrl } = await import("@/app/services/pricingPaymentService");
                const { data } = await getManageSubUrl(apiKey);
                if (data?.manageUrl) window.open(data.manageUrl, "_blank");
              }}
              className="text-sm text-[#00E3AA] hover:underline transition-colors"
            >
              Manage subscription &rarr;
            </button>
          </motion.div>
        )}

        <ContactUsModal open={contactModalOpen} onClose={() => setContactModalOpen(false)} />
      </div>
    </div>
  );
}

// ─── Exported component ───────────────────────────────────────────────────────
export function SubscriptionView() {
  return (
    <SubscriptionContextProvider>
      <SubscriptionViewInner />
    </SubscriptionContextProvider>
  );
}
