import { HeroSection } from "@/components/landing/HeroSection";
import { HowItWorksSection } from "@/components/landing/HowItWorksSection";
import { FeaturesSection } from "@/components/landing/FeaturesSection";
import { AvatarsShowcase } from "@/components/landing/AvatarsShowcase";
import { ApiIntegration } from "@/components/landing/ApiIntegration";
import { PricingSection } from "@/components/landing/PricingSection";
import { PerformanceMetrics } from "@/components/landing/PerformanceMetrics";
import { Footer } from "@/components/landing/Footer";
import { Nav } from "@/components/landing/Nav";
import { stackServerApp } from "@/stack";
import { redirect } from "next/navigation";

export default async function LandingPage() {
  const user = await stackServerApp.getUser({ or: "return-null" });
  if (user) {
    redirect("/dashboard");
  }

  return (
    <div className="min-h-screen bg-[#050505] text-[#E0E0E0] font-inter selection:bg-[#00E3AA] selection:text-black overflow-x-hidden">
      <Nav />
      <main className="flex flex-col gap-0 pb-20">
        <HeroSection />
        <HowItWorksSection />
        <FeaturesSection />
        <AvatarsShowcase />
        <ApiIntegration />
        <PricingSection />
        <PerformanceMetrics />
      </main>
      <Footer />
    </div>
  );
}
