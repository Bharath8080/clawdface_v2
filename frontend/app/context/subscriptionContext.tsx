"use client";

import {
  getLicenseDetails,
  getPricingPlans,
  ILicenseInfo,
  PlanType,
} from "@/app/services/pricingPaymentService";
import { createContext, useCallback, useEffect, useState } from "react";

export type TBillingCycle = "year" | "quarter" | "month";
export type TStepType = "plan" | "summary";

interface SubscriptionContextType {
  allPricingPlans: PlanType[];
  pricingPlans: PlanType[];
  selectedPlan: string;
  setSelectedPlan: React.Dispatch<React.SetStateAction<string>>;
  billingCycle: TBillingCycle;
  setBillingCycle: React.Dispatch<React.SetStateAction<TBillingCycle>>;
  stepType: TStepType;
  setStepType: React.Dispatch<React.SetStateAction<TStepType>>;
  licenseInfo: ILicenseInfo;
  setLicenseInfo: React.Dispatch<React.SetStateAction<ILicenseInfo>>;
  manageSubUrl: string;
  setManageSubUrl: React.Dispatch<React.SetStateAction<string>>;
  isLoading: boolean;
}

const SubscriptionContext = createContext<SubscriptionContextType>({
  allPricingPlans: [],
  pricingPlans: [],
  selectedPlan: "",
  setSelectedPlan: () => {},
  billingCycle: "month",
  setBillingCycle: () => {},
  stepType: "plan",
  setStepType: () => {},
  licenseInfo: {} as ILicenseInfo,
  setLicenseInfo: () => {},
  manageSubUrl: "",
  setManageSubUrl: () => {},
  isLoading: false,
});

const SubscriptionContextProvider = ({ children }: { children: React.ReactNode }) => {
  const [trugenPlans, setTrugenPlans] = useState<PlanType[]>([]);
  const [pricingPlans, setPricingPlans] = useState<PlanType[]>([]);
  const [selectedPlan, setSelectedPlan] = useState<string>("");
  const [billingCycle, setBillingCycle] = useState<TBillingCycle>("month");
  const [stepType, setStepType] = useState<TStepType>("plan");
  const [licenseInfo, setLicenseInfo] = useState<ILicenseInfo>({} as ILicenseInfo);
  const [manageSubUrl, setManageSubUrl] = useState("");
  const [isLoading, setIsLoading] = useState(false);

  const getApiKey = () => {
    if (typeof window === "undefined") return null;
    return localStorage.getItem("defaultApiKey") || null;
  };

  const fetchPricingPlans = useCallback(async () => {
    const apiKey = getApiKey();
    if (!apiKey) return;

    setIsLoading(true);
    try {
      let subPlans: PlanType[] = [];

      const { data: plansData, error: plansError } = await getPricingPlans(apiKey);
      if (plansError) console.error("Error fetching subscription plans:", plansError);
      if (plansData) subPlans = plansData;

      let license: ILicenseInfo = {} as ILicenseInfo;
      const { data: licenseData, error: licenseError } = await getLicenseDetails(apiKey);
      if (licenseError) console.error("Error fetching license details:", licenseError);
      if (licenseData) license = licenseData;

      setBillingCycle(
        license?.slug && license.slug.includes("year") && !license.slug.includes("free")
          ? "year"
          : "month"
      );

      const currentPlanIndex = subPlans.findIndex(
        (p) => p.slug === license?.slug || (p.name === "Free" && !license?.slug)
      );
      if (currentPlanIndex !== -1) {
        subPlans[currentPlanIndex].isCurrent = true;
      }

      setTrugenPlans(subPlans);
      updatePricingPlans(subPlans, license, billingCycle);
      setLicenseInfo(license);
    } catch (error) {
      console.error("Error loading subscription data:", error);
    } finally {
      setIsLoading(false);
    }
  }, []);

  const updatePricingPlans = (
    plans: PlanType[],
    license: ILicenseInfo,
    cycle: TBillingCycle
  ) => {
    let updated = plans
      .filter((p) => p.billingCycle === cycle)
      .sort((a, b) => a.price - b.price);

    if (license?.slug) {
      updated = updated.filter((p) => p.name !== "Free");
    }

    const enterprisePlan = updated.find((p) => p.slug?.includes("ente_ente_"));
    if (enterprisePlan) {
      updated.splice(updated.indexOf(enterprisePlan), 1);
      updated.push(enterprisePlan);
    }

    if (!updated.find((p) => p.slug?.includes("free"))) {
      const freePlan = plans.find((p) => p.slug?.includes("free"));
      if (freePlan) updated.unshift(freePlan);
    }

    setPricingPlans(updated);
  };

  useEffect(() => {
    fetchPricingPlans();
  }, [fetchPricingPlans]);

  useEffect(() => {
    updatePricingPlans(trugenPlans, licenseInfo, billingCycle);
  }, [billingCycle]);

  return (
    <SubscriptionContext.Provider
      value={{
        allPricingPlans: trugenPlans,
        pricingPlans,
        selectedPlan,
        setSelectedPlan,
        billingCycle,
        setBillingCycle,
        stepType,
        setStepType,
        licenseInfo,
        setLicenseInfo,
        manageSubUrl,
        setManageSubUrl,
        isLoading,
      }}
    >
      {children}
    </SubscriptionContext.Provider>
  );
};

export { SubscriptionContextProvider, SubscriptionContext };
