import { Nav } from "@/components/landing/Nav";
import { HeroSection } from "@/components/landing/HeroSection";
import { DemoSection } from "@/components/landing/DemoSection";
import { WhyFaceMattersSection } from "@/components/landing/WhyFaceMattersSection";
import { PlatformsSection } from "@/components/landing/PlatformsSection";
import { HowItWorksSection } from "@/components/landing/HowItWorksSection";
import { WhatYouGetSection } from "@/components/landing/WhatYouGetSection";
import { FeaturesSection } from "@/components/landing/FeaturesSection";
import { BeforeAfterSection } from "@/components/landing/BeforeAfterSection";
import { SecuritySection } from "@/components/landing/SecuritySection";
import { PricingSection } from "@/components/landing/PricingSection";
import { FAQSection } from "@/components/landing/FAQSection";
import { ContactCTASection } from "@/components/landing/ContactCTASection";
import { Footer } from "@/components/landing/Footer";
import { stackServerApp } from "@/stack";
import { redirect } from "next/navigation";

export default async function LandingPage() {
  const user = await stackServerApp.getUser({ or: "return-null" });
  if (user) {
    redirect("/dashboard");
  }

  return (
    <div className="min-h-screen bg-canvas text-[#E0E0E0] selection:bg-brand selection:text-black overflow-x-hidden">
      <Nav />
      <main className="flex flex-col gap-0">
        <HeroSection />
        <DemoSection />
        {/* ── Blended ambient zone ────────────────────────────────────── */}
        <div className="relative bg-canvas overflow-hidden">
          {/* Emerald — top left */}
          <div className="absolute w-72 h-96 left-[-100px] top-[3%] bg-emerald-400 rounded-full blur-[220px] opacity-55 pointer-events-none" />
          {/* Emerald — mid left */}
          <div className="absolute w-56 h-80 left-[-60px] top-[32%] bg-emerald-400 rounded-full blur-[260px] opacity-40 pointer-events-none" />
          {/* Emerald — lower left */}
          <div className="absolute w-48 h-64 left-[-40px] top-[60%] bg-emerald-500 rounded-full blur-[240px] opacity-30 pointer-events-none" />
          {/* Red — mid right */}
          <div className="absolute w-64 h-80 right-[-90px] top-[45%] bg-red-500 rounded-full blur-[210px] opacity-50 pointer-events-none" />
          {/* Red — lower right */}
          <div className="absolute w-48 h-64 right-[-50px] top-[72%] bg-red-600 rounded-full blur-[200px] opacity-30 pointer-events-none" />
          <WhyFaceMattersSection />
          <PlatformsSection />
          <HowItWorksSection />
          <WhatYouGetSection />
          <FeaturesSection />
        </div>
        <BeforeAfterSection />
        <SecuritySection />
        <PricingSection />
        <FAQSection />
        <ContactCTASection />
      </main>
      <Footer />
    </div>
  );
}
