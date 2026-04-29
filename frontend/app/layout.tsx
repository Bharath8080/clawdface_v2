import "@livekit/components-styles";
import { Metadata } from "next";
import { Suspense, ReactNode } from "react";
import { Inter, Outfit, JetBrains_Mono } from "next/font/google";
import "./globals.css";
import { Providers } from "@/components/Providers";
import { StackProvider, StackTheme } from "@stackframe/stack";
import { stackServerApp } from "../stack";

const inter = Inter({
  weight: ["400", "500", "600", "700", "800", "900"],
  subsets: ["latin"],
  variable: "--font-inter",
  display: "swap",
});

const outfit = Outfit({
  subsets: ["latin"],
  variable: "--font-outfit",
});

const jetbrainsMono = JetBrains_Mono({
  subsets: ["latin"],
  variable: "--font-jetbrains-mono",
});

export const metadata: Metadata = {
  title: "ClawdFace",
  icons: {
    icon: [
      { url: "/clawdface-logo.svg", type: "image/svg+xml", sizes: "any" },
      { url: "/logo.png", type: "image/png", sizes: "64x64" },
    ],
    shortcut: { url: "/clawdface-logo.svg", sizes: "any" },
    apple: { url: "/logo.png", sizes: "180x180" },
  }
};

export default function RootLayout({
  children,
}: Readonly<{
  children: ReactNode;
}>) {
  return (
    <html lang="en" className={`dark h-full ${inter.variable} ${outfit.variable} ${jetbrainsMono.variable} ${inter.className}`}>
      <body className="h-full">
        <StackProvider app={stackServerApp}>
          <StackTheme>
            <Providers>
              <Suspense fallback={null}>
                {children}
              </Suspense>
            </Providers>
          </StackTheme>
        </StackProvider>
      </body>
    </html>
  );
}
