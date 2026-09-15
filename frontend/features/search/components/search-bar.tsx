"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { Search as SearchIcon } from "lucide-react";

import { useRouter } from "@/i18n/navigation";
import { Input } from "@/components/ui/input";

export function SearchBar({
  initialValue = "",
  onChange,
}: {
  initialValue?: string;
  /** When provided, the search page drives live results as the user types. */
  onChange?: (value: string) => void;
}) {
  const t = useTranslations("home");
  const router = useRouter();
  const [value, setValue] = useState(initialValue);

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        if (!onChange && value.trim()) {
          router.push(`/search?q=${encodeURIComponent(value.trim())}`);
        }
      }}
      className="relative w-full max-w-xl"
    >
      <SearchIcon className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
      <Input
        value={value}
        onChange={(e) => {
          setValue(e.target.value);
          onChange?.(e.target.value);
        }}
        placeholder={t("searchPlaceholder")}
        className="pl-9"
        autoFocus={!!onChange}
      />
    </form>
  );
}
