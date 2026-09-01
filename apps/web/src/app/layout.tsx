import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";

import { NetworkWarning } from "@/components/layout/network-warning";
import { MobileBottomNav } from "@/components/layout/mobile-bottom-nav";
import { SiteHeader } from "@/components/layout/site-header";
import { TestnetBanner } from "@/components/layout/testnet-banner";

import { Providers } from "./providers";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "SignalArc — Prediction Market Infrastructure on Arc",
  description:
    "Arc-native infrastructure for USDC-settled prediction markets, market intelligence, and resolver workflows.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={`dark ${geistSans.variable} ${geistMono.variable} h-full antialiased`}
    >
      <body className="min-h-full flex flex-col bg-background text-foreground">
        <Providers>
          <TestnetBanner />
          <SiteHeader />
          <NetworkWarning />
          <main className="flex-1 pb-20 md:pb-0">{children}</main>
          <MobileBottomNav />
        </Providers>
      </body>
    </html>
  );
}
