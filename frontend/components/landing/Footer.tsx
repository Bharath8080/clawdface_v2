import Link from "next/link";
import Image from "next/image";

export function Footer() {
  return (
    <footer className="bg-canvas border-t border-white/5 overflow-hidden">
      {/* Top links row */}
      <div className="max-w-7xl mx-auto px-6 py-14">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-10">
          <div className="col-span-1 md:col-span-2">
            <Link href="/" className="flex items-center gap-2 mb-5 hover:opacity-80 transition-opacity w-fit">
              <div className="w-9 h-9 relative flex-shrink-0">
                <Image src="/openclaw.png" alt="ClawdFace" fill className="object-contain" />
              </div>
              <span className="font-bold text-xl text-zinc-400 line-through decoration-2">Clawd</span>
              <span className="font-bold text-xl text-red-500">Face</span>
            </Link>
            <p className="text-zinc-500 text-sm max-w-xs leading-relaxed">
              The bridge between your OpenClaw logic and lifelike interactive avatars. Professional
              video presence for any AI agent.
            </p>
          </div>

          <div>
            <h4 className="text-white font-semibold mb-5 text-sm uppercase tracking-wider">Product</h4>
            <ul className="space-y-3">
              <li>
                <Link href="/#features" className="text-zinc-500 text-sm hover:text-brand transition-colors">
                  Features
                </Link>
              </li>
              <li>
                <Link href="/#pricing" className="text-zinc-500 text-sm hover:text-brand transition-colors">
                  Pricing
                </Link>
              </li>
              <li>
                <Link href="/#demo" className="text-zinc-500 text-sm hover:text-brand transition-colors">
                  Live Demo
                </Link>
              </li>
            </ul>
          </div>

          <div>
            <h4 className="text-white font-semibold mb-5 text-sm uppercase tracking-wider">Legal</h4>
            <ul className="space-y-3">
              <li>
                <Link href="/privacy-policy" className="text-zinc-500 text-sm hover:text-brand transition-colors">
                  Privacy Policy
                </Link>
              </li>
              <li>
                <Link href="/terms-of-service" className="text-zinc-500 text-sm hover:text-brand transition-colors">
                  Terms of Service
                </Link>
              </li>
              <li>
                <Link href="/#contact" className="text-zinc-500 text-sm hover:text-brand transition-colors">
                  Contact Us
                </Link>
              </li>
            </ul>
          </div>
        </div>

        <div className="mt-12 pt-8 border-t border-white/5 flex flex-col md:flex-row items-center justify-between gap-4 text-zinc-600 text-xs">
          <div className="flex items-center gap-2">
            <span className="w-1.5 h-1.5 rounded-full bg-brand animate-pulse" />
            <span>All systems operational</span>
          </div>
          <span>&copy; {new Date().getFullYear()} ClawdFace. All rights reserved.</span>
        </div>
      </div>

      {/* Large bottom branding */}
      <div className="border-t border-white/5 px-6 pb-4 pt-6 flex items-end justify-between gap-4 overflow-hidden">
        <div
          className="text-[clamp(3rem,10vw,9rem)] font-black text-white/5 leading-none select-none tracking-tight whitespace-nowrap"
          aria-hidden="true"
        >
          ClawdFace
        </div>
        <div className="flex-shrink-0 w-20 h-20 relative opacity-20 mb-2">
          <Image src="/openclaw.png" alt="" fill className="object-contain" aria-hidden="true" />
        </div>
      </div>
    </footer>
  );
}
