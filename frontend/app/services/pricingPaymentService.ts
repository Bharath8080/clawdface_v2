export interface IPlanDetails {
  info: string;
  order: number;
  title: string;
  description?: string;
  features: string[];
  creditConsumptionPerMinute: number;
}

export interface PlanType {
  id: string;
  name: string;
  displayName: string;
  description: string;
  userType: string;
  billingCycle: string;
  price: number;
  credits: number;
  billingType: string;
  slug: string;
  valueScore: number;
  concurrentSessions: number;
  maxSessionDuration: number;
  details: IPlanDetails[];
  isCurrent?: boolean;
}

export interface ILicenseInfo {
  id: string;
  organizationId: string;
  concurrentSessions: number;
  totalCredit: number;
  balanceCredit: number;
  purchasedCredit: number;
  createdAt: string;
  expiresAt: string;
  updatedAt: string;
  maxSessionDuration: number;
  subscriptionId: string;
  stripeId: string;
  slug: string;
  cancelledAt: string | null;
  billingCycle: "year" | "quarter" | "month";
  autoReload: boolean;
  autoReloadSlug: string;
}

export interface IBillBreakupInfo {
  daysLeft: number;
  existingPlanRemainingAmount: number;
  targetPlanPrice: number;
  amountToPay: number;
}

export interface IInvoiceInfo {
  id: string;
  organizationId: string;
  createdAt: string;
  creditValue: number;
  displayName: string;
  paymentId: string;
  userId: string;
  price: number;
  subscriptionId: string;
  slug: string;
  planName: string;
  currentPeriodEnd: string | null;
  invoiceURL: string;
}

export const getPublicPricingPlans = async (): Promise<{
  data: PlanType[] | null;
  error: string | null;
}> => {
  try {
    const response = await fetch(`${process.env.NEXT_PUBLIC_BASE_URL}/v1/ext/purchase/plans`, {
      method: "GET",
    });
    if (response.ok) {
      const data = (await response.json()) as PlanType[];
      return { data, error: null };
    }
    const errorText = await response.text();
    return { data: null, error: errorText || "Failed to fetch plans" };
  } catch (err: unknown) {
    const errorMessage = err instanceof Error ? err.message : "Unknown error";
    console.error("Error fetching public pricing plans:", err);
    return { data: null, error: errorMessage };
  }
};

export const getPricingPlans = async (
  apiKey: string
): Promise<{ data: PlanType[] | null; status?: number; error: string | null }> => {
  try {
    const response = await fetch(`${process.env.NEXT_PUBLIC_BASE_URL}/v1/purchase/plans`, {
      headers: { "X-API-Key": apiKey },
      method: "GET",
    });
    if (response.ok) {
      const data = (await response.json()) as PlanType[];
      return { data, error: null };
    } else {
      const errorText = await response.text();
      let errorMessage = "An error occurred";
      try {
        const errorJson = JSON.parse(errorText);
        errorMessage = errorJson.error || errorJson.message || errorText;
      } catch {
        errorMessage = errorText || "An error occurred";
      }
      return { data: null, status: response.status, error: errorMessage };
    }
  } catch (err: unknown) {
    const errorMessage = err instanceof Error ? err.message : "Unknown error";
    console.error("Error fetching pricing plans:", err);
    return { data: null, error: errorMessage };
  }
};

export const getLicenseDetails = async (
  apiKey: string
): Promise<{ data: ILicenseInfo | null; status?: number; error: string | null }> => {
  try {
    const response = await fetch(`${process.env.NEXT_PUBLIC_BASE_URL}/v1/purchase/details`, {
      headers: { "X-API-Key": apiKey },
      method: "GET",
    });
    if (response.ok) {
      const data = (await response.json()) as ILicenseInfo;
      return { data, error: null };
    } else {
      const errorText = await response.text();
      let errorMessage = "An error occurred";
      try {
        const errorJson = JSON.parse(errorText);
        errorMessage = errorJson.error || errorJson.message || errorText;
      } catch {
        errorMessage = errorText || "An error occurred";
      }
      return { data: null, status: response.status, error: errorMessage };
    }
  } catch (err: unknown) {
    const errorMessage = err instanceof Error ? err.message : "Unknown error";
    console.error("Error fetching license details:", err);
    return { data: null, error: errorMessage };
  }
};

export const updateAutoReload = async (
  apiKey: string,
  body: { autoReload: boolean; autoReloadSlug: string }
): Promise<{ data: unknown | null; status?: number; error: string | null }> => {
  try {
    const response = await fetch(`${process.env.NEXT_PUBLIC_BASE_URL}/v1/purchase/auto-reload`, {
      headers: { "X-API-Key": apiKey, "Content-Type": "application/json" },
      method: "POST",
      body: JSON.stringify(body),
    });
    if (response.ok) {
      const data = await response.json();
      return { data, error: null };
    } else {
      const errorText = await response.text();
      let errorMessage = "An error occurred";
      try {
        const errorJson = JSON.parse(errorText);
        errorMessage = errorJson.error || errorJson.message || errorText;
      } catch {
        errorMessage = errorText || "An error occurred";
      }
      return { data: null, status: response.status, error: errorMessage };
    }
  } catch (err: unknown) {
    const errorMessage = err instanceof Error ? err.message : "Unknown error";
    console.error("Error updating auto reload:", err);
    return { data: null, error: errorMessage };
  }
};

export const purchasePlan = async (
  apiKey: string,
  body: object
): Promise<{ data: { checkoutUrl: string } | null; status?: number; error: string | null }> => {
  try {
    const response = await fetch(`${process.env.NEXT_PUBLIC_BASE_URL}/v1/purchase`, {
      headers: { "X-API-Key": apiKey, "Content-Type": "application/json" },
      method: "POST",
      body: JSON.stringify(body),
    });
    if (response.ok) {
      const data = (await response.json()) as { checkoutUrl: string };
      return { data, error: null };
    } else {
      const errorText = await response.text();
      let errorMessage = "An error occurred";
      try {
        const errorJson = JSON.parse(errorText);
        errorMessage = errorJson.error || errorJson.message || errorText;
      } catch {
        errorMessage = errorText || "An error occurred";
      }
      return { data: null, status: response.status, error: errorMessage };
    }
  } catch (err: unknown) {
    const errorMessage = err instanceof Error ? err.message : "Unknown error";
    console.error("Error purchasing subscription:", err);
    return { data: null, error: errorMessage };
  }
};

export const getBillDetails = async (
  apiKey: string,
  body: object
): Promise<{ data: IBillBreakupInfo | null; status?: number; error: string | null }> => {
  try {
    const response = await fetch(`${process.env.NEXT_PUBLIC_BASE_URL}/v1/purchase/plan-price`, {
      headers: { "X-API-Key": apiKey },
      method: "POST",
      body: JSON.stringify(body),
    });
    if (response.ok) {
      const data = (await response.json()) as IBillBreakupInfo;
      return { data, error: null };
    } else {
      const errorText = await response.text();
      let errorMessage = "An error occurred";
      try {
        const errorJson = JSON.parse(errorText);
        errorMessage = errorJson.error || errorJson.message || errorText;
      } catch {
        errorMessage = errorText || "An error occurred";
      }
      return { data: null, status: response.status, error: errorMessage };
    }
  } catch (err: unknown) {
    const errorMessage = err instanceof Error ? err.message : "Unknown error";
    console.error("Error fetching bill details:", err);
    return { data: null, error: errorMessage };
  }
};

export const getPaymentHistory = async (
  apiKey: string
): Promise<{ data: IInvoiceInfo[] | null; status?: number; error: string | null }> => {
  try {
    const response = await fetch(`${process.env.NEXT_PUBLIC_BASE_URL}/v1/purchase/history`, {
      headers: { "X-API-Key": apiKey },
      method: "GET",
    });
    if (response.ok) {
      const data = (await response.json()) as IInvoiceInfo[];
      return { data, error: null };
    } else {
      const errorText = await response.text();
      let errorMessage = "An error occurred";
      try {
        const errorJson = JSON.parse(errorText);
        errorMessage = errorJson.error || errorJson.message || errorText;
      } catch {
        errorMessage = errorText || "An error occurred";
      }
      return { data: null, status: response.status, error: errorMessage };
    }
  } catch (err: unknown) {
    const errorMessage = err instanceof Error ? err.message : "Unknown error";
    console.error("Error fetching payment history:", err);
    return { data: null, error: errorMessage };
  }
};

export interface IContactUsPayload {
  first_name: string;
  last_name: string;
  email: string;
  company: string;
  country: string;
  description?: string;
  intrested_in: string;
  get_notify: boolean;
}

export const contactUsQuery = async (
  body: IContactUsPayload
): Promise<{ data: unknown | null; status?: number; error: string | null }> => {
  try {
    const response = await fetch("https://infra.trugen.ai/landing/contact", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    if (response.ok) {
      const data = await response.json();
      return { data, error: null };
    } else {
      const errorText = await response.text();
      let errorMessage = "An error occurred";
      try {
        const errorJson = JSON.parse(errorText);
        errorMessage = errorJson.error || errorJson.message || errorText;
      } catch {
        errorMessage = errorText || "An error occurred";
      }
      return { data: null, status: response.status, error: errorMessage };
    }
  } catch (err: unknown) {
    const errorMessage = err instanceof Error ? err.message : "Unknown error";
    console.error("Error posting contact us query:", err);
    return { data: null, error: errorMessage };
  }
};

export const getManageSubUrl = async (
  apiKey: string
): Promise<{ data: { manageUrl: string } | null; status?: number; error: string | null }> => {
  try {
    const response = await fetch(`${process.env.NEXT_PUBLIC_BASE_URL}/v1/purchase/manage-subscription`, {
      headers: { "X-API-Key": apiKey },
      method: "GET",
    });
    if (response.ok) {
      const data = (await response.json()) as { manageUrl: string };
      return { data, error: null };
    } else {
      const errorText = await response.text();
      let errorMessage = "An error occurred";
      try {
        const errorJson = JSON.parse(errorText);
        errorMessage = errorJson.error || errorJson.message || errorText;
      } catch {
        errorMessage = errorText || "An error occurred";
      }
      return { data: null, status: response.status, error: errorMessage };
    }
  } catch (err: unknown) {
    const errorMessage = err instanceof Error ? err.message : "Unknown error";
    console.error("Error fetching manage subscription URL:", err);
    return { data: null, error: errorMessage };
  }
};
