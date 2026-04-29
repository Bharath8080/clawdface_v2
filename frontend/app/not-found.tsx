import Link from "next/link";

export default function NotFound() {
  return (
    <div className="min-h-screen bg-canvas flex items-center justify-center p-4">
      <div className="flex flex-col items-center gap-6 text-center max-w-md">
        <div className="relative">
          <span className="text-[120px] font-black text-white/[0.04] leading-none select-none">
            404
          </span>
          <div className="absolute inset-0 flex items-center justify-center">
            <div className="w-16 h-16 bg-brand/10 border border-brand/20 rounded-2xl flex items-center justify-center">
              <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#00E3AA" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
                <circle cx="11" cy="11" r="8" />
                <path d="m21 21-4.35-4.35" />
                <path d="M11 8v3M11 14h.01" />
              </svg>
            </div>
          </div>
        </div>

        <div className="flex flex-col gap-2">
          <h1 className="text-2xl font-bold text-white tracking-tight">Page not found</h1>
          <p className="text-[#6b7280] text-sm leading-relaxed">
            The page you&apos;re looking for doesn&apos;t exist or has been moved.
          </p>
        </div>

        <div className="flex items-center gap-3">
          <Link
            href="/dashboard"
            className="px-6 py-2.5 bg-brand text-black text-sm font-bold rounded-xl hover:brightness-110 transition-all active:scale-[0.98]"
          >
            Go to Dashboard
          </Link>
          <Link
            href="/log-in"
            className="px-6 py-2.5 bg-white/[0.06] border border-white/[0.1] text-white text-sm font-semibold rounded-xl hover:bg-white/[0.1] transition-all active:scale-[0.98]"
          >
            Sign In
          </Link>
        </div>
      </div>
    </div>
  );
}
