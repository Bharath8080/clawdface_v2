import "@livekit/components-styles";
import { Metadata } from "next";
import { Inter, Outfit, JetBrains_Mono } from "next/font/google";
import "./globals.css";
import { Providers } from "@/components/Providers";
import { StackProvider, StackTheme } from "@stackframe/stack";
import { stackServerApp } from "../stack";

const inter = Inter({
  weight: ["400", "500", "600", "700"],
  subsets: ["latin"],
  variable: "--font-inter",
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
    icon: "/openclaw.png",
    shortcut: "/openclaw.png",
    apple: "/openclaw.png",
  }
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className={`dark h-full ${inter.variable} ${outfit.variable} ${jetbrainsMono.variable} ${inter.className}`}>
      <body className="h-full">
        <StackProvider app={stackServerApp}>
          <StackTheme>
            <Providers>{children}</Providers>
          </StackTheme>
        </StackProvider>
      </body>
    </html>
  );
}
