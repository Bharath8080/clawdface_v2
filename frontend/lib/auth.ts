"use client";

export const THEME_KEY = "clawdface_theme";

export function getTheme(): "dark" | "light" {
  if (typeof window === "undefined") return "dark";
  return (localStorage.getItem(THEME_KEY) as "dark" | "light") || "dark";
}

export function setTheme(theme: "dark" | "light") {
  if (typeof window === "undefined") return;
  localStorage.setItem(THEME_KEY, theme);
  if (theme === "light") {
    document.documentElement.classList.add("light");
  } else {
    document.documentElement.classList.remove("light");
  }
  window.dispatchEvent(new CustomEvent("clawdface-theme-change", { detail: theme }));
}

export function getInitials(email?: string | null, name?: string | null): string {
  if (!email && !name) return "?";
  if (name) {
    const parts = name.trim().split(" ");
    return parts.slice(0, 2).map((p) => p[0]).join("").toUpperCase();
  }
  return email ? email.slice(0, 2).toUpperCase() : "?";
}
