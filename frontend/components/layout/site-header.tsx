"use client";

import { useTranslations } from "next-intl";
import { BookOpen, Heart, Search as SearchIcon, ShieldCheck } from "lucide-react";

import { Link } from "@/i18n/navigation";
import { buttonVariants } from "@/components/ui/button";
import { LocaleSwitcher } from "@/components/layout/locale-switcher";
import { ThemeToggle } from "@/components/layout/theme-toggle";
import { UserMenu } from "@/components/layout/user-menu";
import { useAuthStore } from "@/store/auth-store";
import { cn } from "@/lib/utils";

export function SiteHeader() {
  const t = useTranslations("nav");
  const user = useAuthStore((s) => s.user);
  const canManage = user?.role === "admin" || user?.role === "editor";

  // Nav items are real navigation, not actions — they get button-like styling
  // via `buttonVariants` applied straight to a `Link`, not the Button
  // component. Wrapping a Link in Button (Base UI) exposes it with
  // role="button" (Base UI always applies button a11y semantics, even to a
  // rendered <a>), which is the wrong announcement/keyboard behavior for
  // something screen readers and "open in new tab" should treat as a link.
  const navLinkClass = cn(buttonVariants({ variant: "ghost", size: "sm" }));

  return (
    <header className="sticky top-0 z-40 border-b bg-background/95 backdrop-blur supports-backdrop-filter:bg-background/60">
      <div className="mx-auto flex h-14 max-w-7xl items-center gap-4 px-4">
        <Link href="/" className="flex items-center gap-2 font-semibold">
          <BookOpen className="size-5" />
          <span>Book Reader</span>
        </Link>

        <nav className="hidden md:flex items-center gap-1 text-sm">
          <Link href="/" className={navLinkClass}>
            {t("home")}
          </Link>
          <Link href="/search" className={navLinkClass}>
            <SearchIcon className="size-4" />
            {t("search")}
          </Link>
          {user && (
            <Link href="/favorites" className={navLinkClass}>
              <Heart className="size-4" />
              {t("favorites")}
            </Link>
          )}
          {canManage && (
            <Link href="/admin" className={navLinkClass}>
              <ShieldCheck className="size-4" />
              {t("admin")}
            </Link>
          )}
        </nav>

        <div className="ml-auto flex items-center gap-2">
          <LocaleSwitcher />
          <ThemeToggle />
          <UserMenu />
        </div>
      </div>
    </header>
  );
}
