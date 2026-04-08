import Link from "next/link";

export function Footer() {
  return (
    <footer id="contact" className="py-16 px-6 border-t border-white/5 bg-[#050505] relative overflow-hidden">
      {/* Subtle background glow to keep it from looking "diminished" */}
      <div className="absolute top-0 left-1/2 -translate-x-1/2 w-[600px] h-[300px] bg-[#00E3AA]/5 rounded-full blur-[100px] pointer-events-none" />

      <div className="max-w-7xl mx-auto w-full relative z-10">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-12 mb-16">
          <div className="col-span-1 md:col-span-2">
            <Link href="/" className="flex items-center gap-3 mb-6 hover:opacity-80 transition-opacity w-fit">
              <svg width="28" height="28" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M12 2C12 2 15 6 15 9C15 11 14 12 14 12C14 12 17 9 18 10C19 11 20 13 20 15C20 19.4183 16.4183 23 12 23C7.58172 23 4 19.4183 4 15C4 11 8 6 12 2Z" fill="#00E3AA"/>
              </svg>
              <span className="font-outfit font-bold text-2xl text-white tracking-tight">ClawdFace</span>
            </Link>
            <p className="text-zinc-300 text-base max-w-sm leading-relaxed">
              The bridge between your OpenClaw logic and lifelike interactive avatars. Professional video presence for any AI agent.
            </p>
          </div>

          <div>
            <h4 className="text-white font-bold mb-6">Product</h4>
            <ul className="space-y-4">
              <li><Link href="/#features" className="text-zinc-300 text-base hover:text-[#00E3AA] transition-colors">Features</Link></li>
              <li><Link href="/#pricing" className="text-zinc-300 text-base hover:text-[#00E3AA] transition-colors">Pricing</Link></li>
              <li><Link href="/#demo" className="text-zinc-300 text-base hover:text-[#00E3AA] transition-colors">Live Demo</Link></li>
            </ul>
          </div>

          <div>
            <h4 className="text-white font-bold mb-6">Legal</h4>
            <ul className="space-y-4">
              <li><Link href="/privacy-policy" className="text-zinc-300 text-base hover:text-[#00E3AA] transition-colors">Privacy Policy</Link></li>
              <li><Link href="/terms-of-service" className="text-zinc-300 text-base hover:text-[#00E3AA] transition-colors">Terms of Service</Link></li>
              <li><Link href="/#contact" className="text-zinc-300 text-base hover:text-[#00E3AA] transition-colors">Contact</Link></li>
            </ul>
          </div>
        </div>

        <div className="pt-8 border-t border-white/5 flex flex-col md:flex-row items-center justify-between gap-6 text-zinc-300 text-base">
          <div className="flex items-center gap-2">
            <span className="w-2 h-2 rounded-full bg-[#00E3AA] animate-pulse" />
            <span>Systems Operational</span>
          </div>
          <div>
            &copy; {new Date().getFullYear()} ClawdFace. All rights reserved.
          </div>
        </div>
      </div>
    </footer>
  );
}
