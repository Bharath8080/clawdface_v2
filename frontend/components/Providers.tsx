"use client";
import { useEffect } from "react";
import { getTheme } from "@/lib/auth";

export function Providers({ children }: { children: React.ReactNode }) {
  useEffect(() => {
    // Always enforce dark mode on mount — remove any stale light class
    const theme = getTheme();
    if (theme === "light") {
      document.documentElement.classList.remove("light");
    }
  }, []);

  return (
    <>
      {children}
    </>
  );
}
