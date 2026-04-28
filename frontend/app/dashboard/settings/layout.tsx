"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useUser } from "@stackframe/stack";
import { Sidebar } from "@/components/Sidebar";
import Image from "next/image";

export default function SettingsLayout({ children }: { children: React.ReactNode }) {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const router = useRouter();
  const user = useUser();

  if (user === null) {
    router.replace("/log-in");
    return null;
  }

  if (user && !user.primaryEmailVerified) {
    router.replace("/email-not-verified");
    return null;
  }

  return (
    <main className="h-[100dvh] w-screen bg-canvas flex overflow-hidden font-[Inter] text-white">
      <Sidebar
        activeSession=""
        setActiveSession={() => router.push("/dashboard")}
        isMobileMenuOpen={isMobileMenuOpen}
        setIsMobileMenuOpen={setIsMobileMenuOpen}
      />

      <div className="flex-1 h-full w-full overflow-hidden flex flex-col relative z-0">
        {/* Mobile Header */}
        <div className="md:hidden flex items-center justify-between px-4 h-14 border-b border-white/5 bg-surface-secondary shrink-0 z-10 shadow-sm">
          <div className="flex items-center gap-3">
            <div className="w-7 h-7 shrink-0 relative flex items-center justify-center rounded-lg bg-brand/10 text-brand">
              <Image src="/openclaw.png" alt="Logo" width={18} height={18} className="object-contain drop-shadow-[0_0_4px_rgba(0,227,170,0.5)]" />
            </div>
            <span className="text-white font-bold text-lg leading-none tracking-tight mt-1 font-outfit">ClawdFace</span>
          </div>
          <button
            onClick={() => setIsMobileMenuOpen(true)}
            className="text-white/70 hover:text-white p-2.5 rounded-xl bg-white/[0.03] hover:bg-white/[0.08] transition-all border border-white/5 active:scale-95"
          >
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
              <line x1="3" x2="21" y1="12" y2="12" /><line x1="3" x2="21" y1="6" y2="6" /><line x1="3" x2="21" y1="18" y2="18" />
            </svg>
          </button>
        </div>

        <div className="flex-1 overflow-hidden">
          {children}
        </div>
      </div>
    </main>
  );
}
