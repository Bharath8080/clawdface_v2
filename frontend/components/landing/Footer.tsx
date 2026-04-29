import Link from "next/link";

export function Footer() {
  return (
    <footer className="overflow-hidden bg-gradient-to-b from-zinc-950 to-emerald-400">
      {/* ── Top section ─────────────────────────────────────────────────── */}
      <div className="max-w-[1312px] mx-auto px-6 md:px-10 py-10 md:py-16 flex flex-col gap-10">
        {/* Brand + nav row */}
        <div className="flex flex-col md:flex-row justify-between items-start gap-12">
          {/* Brand */}
          <div className="flex-1 flex flex-col gap-6">
            <Link href="/" className="flex items-center gap-2 w-fit hover:opacity-80 transition-opacity">
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img src="/clawdface-logo.svg" alt="ClawdFace" className="w-14 h-11 object-contain shrink-0" />
              <span className="text-2xl font-semibold capitalize">
                <span className="text-white">Clawd</span>
                <span className="text-emerald-400">Face</span>
              </span>
            </Link>
            <p className="w-96 max-w-full text-neutral-400 text-base font-normal">
              The bridge between your OpenClaw logic and lifelike interactive avatars. Professional video
              presence for any AI agent.
            </p>
          </div>

          {/* Nav columns */}
          <div className="flex gap-12 md:gap-28">
            {/* Product */}
            <div className="flex flex-col gap-4">
              <span className="text-white text-base font-medium">Product</span>
              {[
                { label: "Features",   href: "/#features" },
                { label: "Pricing",    href: "/#pricing"  },
                { label: "Live Demos", href: "/#demo"     },
              ].map((l) => (
                <Link key={l.label} href={l.href} className="text-neutral-400 text-base font-medium hover:text-white transition-colors">
                  {l.label}
                </Link>
              ))}
            </div>

            {/* Legal */}
            <div className="flex flex-col gap-4">
              <span className="text-white text-base font-medium">Legal</span>
              {[
                { label: "Privacy Policy",   href: "/privacy-policy"   },
                { label: "Terms of Service", href: "/terms-of-service" },
                { label: "Contact",          href: "/#contact"         },
              ].map((l) => (
                <Link key={l.label} href={l.href} className="text-neutral-400 text-base font-medium hover:text-white transition-colors">
                  {l.label}
                </Link>
              ))}
            </div>
          </div>
        </div>

        {/* Divider + bottom bar */}
        <div className="flex flex-col gap-10">
          <div className="border-t border-white/10" />
          <div className="flex flex-col md:flex-row justify-between items-center flex-wrap gap-4">
            <div className="p-2 flex items-center gap-2.5">
              <div className="w-1.5 h-1.5 bg-emerald-400 rounded-full animate-pulse" />
              <span className="text-zinc-400 text-base font-medium">Systems Operational</span>
            </div>
            <span className="text-white/80 text-base font-normal">© ClawdFace. All rights reserved.</span>
          </div>
        </div>
      </div>

      {/* ── Mascot + wordmark ───────────────────────────────────────────── */}
      <div className="flex justify-center items-center gap-2 md:gap-10">
        {/* eslint-disable-next-line @next/next/no-img-element */}
        <img src="/clawdface-logo.svg" alt="ClawdFace mascot" className="w-28 h-24 md:w-64 md:h-52 object-contain shrink-0" />
        <span
          className="font-bold text-emerald-400 select-none leading-none whitespace-nowrap"
          style={{ fontSize: "clamp(36px, 10vw, 186px)" }}
          aria-hidden="true"
        >
          ClawdFace
        </span>
      </div>
    </footer>
  );
}
