"use client";

import { useState } from "react";
import { purchasePlan } from "@/app/services/pricingPaymentService";

const usePayment = () => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const paymentHandler = async (planSlug: string) => {
    const apiKey = typeof window !== "undefined"
      ? localStorage.getItem("defaultApiKey")
      : null;

    if (!apiKey) {
      setError("No API key found. Please refresh the page and try again.");
      console.error("No API key found in localStorage.");
      return;
    }

    try {
      setLoading(true);
      setError(null);

      const { data, error: purchaseError } = await purchasePlan(apiKey, {
        subscriptionPlanId: planSlug,
      });

      if (purchaseError) {
        setError(purchaseError);
        console.error("Error purchasing plan:", purchaseError);
        return;
      }

      if (data?.checkoutUrl) {
        window.location.href = data.checkoutUrl;
      }
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : "Unable to process payment. Please contact support.";
      setError(message);
      console.error("Payment error:", err);
    } finally {
      setLoading(false);
    }
  };

  return { paymentHandler, loading, error };
};

export default usePayment;
