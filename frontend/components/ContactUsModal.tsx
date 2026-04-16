"use client";

import { AnimatePresence, motion } from "framer-motion";
import { useState } from "react";
import { contactUsQuery } from "@/app/services/pricingPaymentService";
import countries from "./countries.json";

function isValidEmail(email: string) {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
}

const XIcon = () => (
  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
    <path d="M18 6 6 18M6 6l12 12" />
  </svg>
);

const CheckCircleIcon = () => (
  <svg width="44" height="44" viewBox="0 0 24 24" fill="none" stroke="#00E3AA" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
    <circle cx="12" cy="12" r="10" />
    <polyline points="20 6 9 17 4 12" />
  </svg>
);

const SpinnerIcon = () => (
  <svg className="animate-spin" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
    <path d="M21 12a9 9 0 1 1-6.219-8.56" />
  </svg>
);

interface Field {
  label: string;
  required?: boolean;
  children: React.ReactNode;
}

function Field({ label, required, children }: Field) {
  return (
    <div className="flex flex-col gap-1.5">
      <label className="text-[13px] font-semibold text-[#9ca3af]">
        {label} {required && <span className="text-[#00E3AA]">*</span>}
      </label>
      {children}
    </div>
  );
}

const inputCls =
  "w-full bg-white/[0.04] border border-white/[0.1] rounded-xl px-4 py-2.5 text-[14px] text-white placeholder:text-[#4b5563] focus:outline-none focus:border-[#00E3AA]/40 transition-colors";

export function ContactUsModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName]   = useState("");
  const [email, setEmail]         = useState("");
  const [emailErr, setEmailErr]   = useState(false);
  const [company, setCompany]     = useState("");
  const [country, setCountry]     = useState("");
  const [description, setDescription] = useState("");
  const [loading, setLoading]     = useState(false);
  const [error, setError]         = useState<string | null>(null);
  const [submitted, setSubmitted] = useState(false);

  const reset = () => {
    setFirstName(""); setLastName(""); setEmail(""); setEmailErr(false);
    setCompany(""); setCountry(""); setDescription("");
    setError(null); setSubmitted(false); setLoading(false);
  };

  const handleClose = () => { reset(); onClose(); };

  const disabled = !firstName || !lastName || !email || !company || !country || emailErr;

  const handleSubmit = async () => {
    if (!isValidEmail(email)) { setEmailErr(true); return; }
    setLoading(true);
    setError(null);
    const { data, error: apiError } = await contactUsQuery({
      first_name: firstName,
      last_name: lastName,
      email,
      company,
      country,
      description,
      intrested_in: "Enterprise Plan",
      get_notify: true,
    });
    setLoading(false);
    if (apiError) { setError(apiError); return; }
    if (data !== null) setSubmitted(true);
  };

  return (
    <AnimatePresence>
      {open && (
        <div className="fixed inset-0 z-[200] flex items-center justify-center p-4">
          {/* Backdrop */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="absolute inset-0 bg-black/70 backdrop-blur-sm"
            onClick={handleClose}
          />

          {/* Modal */}
          <motion.div
            initial={{ opacity: 0, scale: 0.96, y: 16 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.96, y: 16 }}
            transition={{ duration: 0.25, ease: [0.22, 1, 0.36, 1] }}
            className="relative w-full max-w-[560px] max-h-[90vh] bg-[#0a0a0a] border border-white/[0.1] rounded-2xl shadow-[0_32px_80px_rgba(0,0,0,0.6)] overflow-hidden flex flex-col"
          >
            {/* Header */}
            <div className="flex items-start justify-between px-7 pt-7 pb-5 shrink-0">
              <div>
                <h2 className="text-xl font-bold text-white tracking-tight">Contact Sales</h2>
                <p className="text-[#6b7280] text-[13px] mt-1">Tell us about your needs and we'll get back to you.</p>
              </div>
              <button onClick={handleClose} className="text-[#4b5563] hover:text-white transition-colors p-1 -mt-0.5 shrink-0">
                <XIcon />
              </button>
            </div>

            <div className="h-px bg-white/[0.06] shrink-0" />

            {/* Content */}
            {submitted ? (
              <div className="flex flex-col items-center justify-center gap-5 py-16 px-7">
                <div className="w-16 h-16 rounded-2xl bg-[#00E3AA]/10 border border-[#00E3AA]/20 flex items-center justify-center">
                  <CheckCircleIcon />
                </div>
                <div className="text-center">
                  <p className="text-white font-bold text-lg">Message sent!</p>
                  <p className="text-[#6b7280] text-sm mt-1">Our team will reach out to you shortly.</p>
                </div>
                <button
                  onClick={handleClose}
                  className="mt-2 px-8 py-2.5 bg-[#00E3AA] text-black text-sm font-bold rounded-xl hover:brightness-110 transition-all active:scale-[0.98]"
                >
                  Done
                </button>
              </div>
            ) : (
              <>
                <div className="overflow-y-auto flex-1 px-7 py-6 custom-scrollbar">
                  <div className="flex flex-col gap-5">
                    {/* Name row */}
                    <div className="grid grid-cols-2 gap-4">
                      <Field label="First Name" required>
                        <input className={inputCls} value={firstName} onChange={e => setFirstName(e.target.value)} placeholder="John" />
                      </Field>
                      <Field label="Last Name" required>
                        <input className={inputCls} value={lastName} onChange={e => setLastName(e.target.value)} placeholder="Doe" />
                      </Field>
                    </div>

                    {/* Email */}
                    <Field label="Work Email" required>
                      <input
                        className={`${inputCls} ${emailErr ? "border-red-500/50 focus:border-red-500/70" : ""}`}
                        type="email"
                        value={email}
                        onChange={e => { setEmail(e.target.value); if (emailErr) setEmailErr(false); }}
                        onBlur={() => { if (email && !isValidEmail(email)) setEmailErr(true); }}
                        placeholder="john@company.com"
                      />
                      {emailErr && <span className="text-red-400 text-[12px] mt-0.5">Please enter a valid email address.</span>}
                    </Field>

                    {/* Company + Country */}
                    <div className="grid grid-cols-2 gap-4">
                      <Field label="Company" required>
                        <input className={inputCls} value={company} onChange={e => setCompany(e.target.value)} placeholder="Acme Inc." />
                      </Field>
                      <Field label="Country" required>
                        <select
                          className={`${inputCls} appearance-none cursor-pointer`}
                          value={country}
                          onChange={e => setCountry(e.target.value)}
                        >
                          <option value="" disabled className="bg-[#111]">Select country</option>
                          {countries.map(c => (
                            <option key={c.code} value={c.name} className="bg-[#111]">{c.name}</option>
                          ))}
                        </select>
                      </Field>
                    </div>

                    {/* Description */}
                    <Field label="Tell us about your use case">
                      <textarea
                        className={`${inputCls} resize-none min-h-[100px]`}
                        value={description}
                        onChange={e => setDescription(e.target.value)}
                        placeholder="Describe your needs, team size, expected usage..."
                      />
                    </Field>

                    {/* API error */}
                    {error && (
                      <div className="p-3 rounded-xl bg-red-500/10 border border-red-500/20 text-red-400 text-[13px]">
                        {error}
                      </div>
                    )}
                  </div>
                </div>

                {/* Footer */}
                <div className="h-px bg-white/[0.06] shrink-0" />
                <div className="flex items-center justify-between gap-3 px-7 py-5 shrink-0">
                  <button onClick={handleClose} className="text-[#6b7280] hover:text-white text-sm font-semibold transition-colors">
                    Cancel
                  </button>
                  <button
                    onClick={handleSubmit}
                    disabled={disabled || loading}
                    className="flex items-center gap-2 px-6 py-2.5 bg-[#00E3AA] text-black text-sm font-bold rounded-xl hover:brightness-110 active:scale-[0.98] transition-all disabled:opacity-40 disabled:cursor-not-allowed disabled:hover:brightness-100"
                  >
                    {loading && <SpinnerIcon />}
                    {loading ? "Sending..." : "Send Message"}
                  </button>
                </div>
              </>
            )}
          </motion.div>
        </div>
      )}
    </AnimatePresence>
  );
}
