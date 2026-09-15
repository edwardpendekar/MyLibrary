import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { NextIntlClientProvider, hasLocale } from "next-intl";
import { Geist, Geist_Mono, Lora } from "next/font/google";

import { routing } from "@/i18n/routing";
import { Providers } from "@/components/providers";
import { SiteHeader } from "@/components/layout/site-header";
import "../globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

// A literary serif for headings and book titles — deliberate "reading app"
// character (à la Kindle/Apple Books), distinct from the sans-serif UI chrome.
const lora = Lora({
  variable: "--font-lora",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: {
    default: "Book Reader",
    template: "%s · Book Reader",
  },
  description: "Read structured, multi-language texts online or as PDF.",
};

export function generateStaticParams() {
  return routing.locales.map((locale) => ({ locale }));
}

export default async function LocaleLayout({
  children,
  params,
}: LayoutProps<"/[locale]">) {
  const { locale } = await params;
  if (!hasLocale(routing.locales, locale)) {
    notFound();
  }

  return (
    <html
      lang={locale}
      suppressHydrationWarning
      className={`${geistSans.variable} ${geistMono.variable} ${lora.variable} h-full antialiased`}
    >
      <body className="min-h-full flex flex-col bg-background text-foreground">
        <NextIntlClientProvider>
          <Providers>
            <SiteHeader />
            <main className="flex-1">{children}</main>
          </Providers>
        </NextIntlClientProvider>
      </body>
    </html>
  );
}
