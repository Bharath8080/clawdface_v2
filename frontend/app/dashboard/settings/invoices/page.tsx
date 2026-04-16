"use client";

import { motion } from "framer-motion";
import { useEffect, useState } from "react";
import { getPaymentHistory } from "@/app/services/pricingPaymentService";

const DownloadIcon = () => (
  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
    <polyline points="7 10 12 15 17 10" />
    <line x1="12" y1="15" x2="12" y2="3" />
  </svg>
);

const ReceiptIcon = () => (
  <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
    <path d="M4 2v20l2-1 2 1 2-1 2 1 2-1 2 1 2-1 2 1V2l-2 1-2-1-2 1-2-1-2 1-2-1-2 1-2-1Z" />
    <path d="M16 8H8" /><path d="M16 12H8" /><path d="M12 16H8" />
  </svg>
);

interface InvoiceRow {
  planName: string;
  paidOn: string;
  amount: number;
  invoiceURL: string;
}

export default function InvoicesPage() {
  const [invoices, setInvoices] = useState<InvoiceRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchInvoices = async () => {
      try {
        const apiKey = localStorage.getItem("defaultApiKey");
        if (!apiKey) {
          setError("No API key found. Please refresh the page.");
          return;
        }

        const { data, error: apiError } = await getPaymentHistory(apiKey);
        if (apiError) {
          setError(apiError);
          return;
        }

        if (data) {
          setInvoices(
            data.map((d) => ({
              planName: d.displayName,
              paidOn: d.createdAt,
              amount: d.price,
              invoiceURL: d.invoiceURL,
            }))
          );
        }
      } catch (err) {
        setError("Failed to load invoices.");
        console.error(err);
      } finally {
        setLoading(false);
      }
    };

    fetchInvoices();
  }, []);

  const formatDate = (dateStr: string) => {
    const d = new Date(dateStr);
    if (isNaN(d.getTime())) return "—";
    return new Intl.DateTimeFormat("en-GB", {
      day: "2-digit",
      month: "short",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
      hour12: true,
    }).format(d);
  };

  const formatAmount = (amount: number) =>
    typeof amount === "number"
      ? amount.toLocaleString("en-US", { style: "currency", currency: "USD" })
      : `$${amount}`;

  return (
    <div className="h-full overflow-y-auto bg-[#050505] custom-scrollbar">
      {/* Background glow */}
      <div className="pointer-events-none fixed inset-0 overflow-hidden">
        <div className="absolute top-1/3 left-1/2 -translate-x-1/2 w-[500px] h-[500px] bg-[#00E3AA]/3 rounded-full blur-[120px]" />
      </div>

      <div className="relative z-10 max-w-4xl mx-auto px-6 py-10 pb-16">

        {/* Header */}
        <motion.div
          initial={{ opacity: 0, y: 16 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4 }}
          className="mb-8"
        >
          <div className="flex items-center gap-2 mb-3">
            <span className="text-[11px] font-bold uppercase tracking-widest text-[#00E3AA] bg-[#00E3AA]/10 px-3 py-1 rounded-full border border-[#00E3AA]/20">
              Billing
            </span>
          </div>
          <h1 className="text-3xl font-bold text-white mb-2 tracking-tight">Invoices</h1>
          <p className="text-[#6b7280] text-[15px]">All your payment receipts in one place.</p>
        </motion.div>

        {/* Loading */}
        {loading && (
          <div className="space-y-3">
            {[1, 2, 3].map((i) => (
              <div key={i} className="h-16 rounded-xl bg-white/[0.02] border border-white/5 animate-pulse" />
            ))}
          </div>
        )}

        {/* Error */}
        {!loading && error && (
          <motion.div
            initial={{ opacity: 0, y: -8 }}
            animate={{ opacity: 1, y: 0 }}
            className="p-4 rounded-xl bg-red-500/10 border border-red-500/20 text-red-400 text-sm"
          >
            {error}
          </motion.div>
        )}

        {/* Empty state */}
        {!loading && !error && invoices.length === 0 && (
          <motion.div
            initial={{ opacity: 0, y: 16 }}
            animate={{ opacity: 1, y: 0 }}
            className="flex flex-col items-center justify-center py-28 gap-5"
          >
            <div className="w-20 h-20 rounded-2xl bg-white/[0.03] border border-white/8 flex items-center justify-center text-[#374151]">
              <ReceiptIcon />
            </div>
            <div className="text-center">
              <p className="text-white font-semibold text-lg">No invoices yet</p>
              <p className="text-[#6b7280] text-sm mt-1">Payment receipts will appear here after your first purchase.</p>
            </div>
          </motion.div>
        )}

        {/* Table */}
        {!loading && !error && invoices.length > 0 && (
          <motion.div
            initial={{ opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.1 }}
            className="overflow-hidden rounded-2xl border border-white/8"
          >
            <table className="min-w-full text-sm text-left">
              <thead>
                <tr className="bg-white/[0.03] border-b border-white/8">
                  <th className="px-5 py-3.5 text-[11px] font-bold uppercase tracking-widest text-[#6b7280]">Plan</th>
                  <th className="px-5 py-3.5 text-[11px] font-bold uppercase tracking-widest text-[#6b7280]">Date</th>
                  <th className="px-5 py-3.5 text-[11px] font-bold uppercase tracking-widest text-[#6b7280]">Amount</th>
                  <th className="px-5 py-3.5 text-[11px] font-bold uppercase tracking-widest text-[#6b7280]">Receipt</th>
                </tr>
              </thead>
              <tbody>
                {invoices.map((inv, idx) => (
                  <tr
                    key={idx}
                    className="border-b border-white/5 last:border-0 hover:bg-white/[0.02] transition-colors"
                  >
                    <td className="px-5 py-4 text-white font-medium">{inv.planName || "—"}</td>
                    <td className="px-5 py-4 text-[#9ca3af]">{formatDate(inv.paidOn)}</td>
                    <td className="px-5 py-4 text-[#00E3AA] font-semibold">{formatAmount(inv.amount)}</td>
                    <td className="px-5 py-4">
                      {inv.invoiceURL ? (
                        <a
                          href={inv.invoiceURL}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="inline-flex items-center gap-2 text-[13px] font-semibold text-[#6b7280] hover:text-white transition-colors border border-white/8 hover:border-white/20 px-3 py-1.5 rounded-lg bg-white/[0.03] hover:bg-white/[0.06] active:scale-95"
                        >
                          <DownloadIcon />
                          Download
                        </a>
                      ) : (
                        <span className="text-[#374151] text-sm">—</span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </motion.div>
        )}
      </div>
    </div>
  );
}
