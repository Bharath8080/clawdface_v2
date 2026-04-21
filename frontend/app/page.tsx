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
    <div className="min-h-screen bg-[#050505] text-[#E0E0E0] selection:bg-[#00E3AA] selection:text-black overflow-x-hidden">
      <Nav />
      <main className="flex flex-col gap-0">
        <HeroSection />
        <DemoSection />
        <WhyFaceMattersSection />
        <PlatformsSection />
        <HowItWorksSection />
        <WhatYouGetSection />
        <FeaturesSection />
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
