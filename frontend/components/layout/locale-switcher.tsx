"use client";

import { useLocale } from "next-intl";

import { routing } from "@/i18n/routing";
import { usePathname, useRouter } from "@/i18n/navigation";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

const LOCALE_LABELS: Record<string, string> = { en: "English", id: "Indonesia" };

export function LocaleSwitcher() {
  const locale = useLocale();
  const router = useRouter();
  const pathname = usePathname();

  return (
    <Select
      value={locale}
      onValueChange={(next) => {
        if (next) router.replace(pathname, { locale: next });
      }}
    >
      <SelectTrigger className="w-[110px]" aria-label="Language">
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        {routing.locales.map((l) => (
          <SelectItem key={l} value={l}>
            {LOCALE_LABELS[l] ?? l}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}
